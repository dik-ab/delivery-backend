package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/delivery-app/delivery-api/internal/util"
	"github.com/gin-gonic/gin"
)

// PredictHandler handles predicted location requests
type PredictHandler struct {
	tripRepo *repository.TripRepository
}

// NewPredictHandler creates a new predict handler
func NewPredictHandler(tripRepo *repository.TripRepository) *PredictHandler {
	return &PredictHandler{tripRepo: tripRepo}
}

// GetPredictedLocation godoc
// @Summary Get predicted location of a trip at current time
// @Description Calculate predicted position based on departure time and route steps
// @Tags predict
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Param at query string false "Target time (RFC3339). Defaults to current time"
// @Success 200 {object} util.PredictedLocation
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/trips/{trip_id}/predict [get]
func (h *PredictHandler) GetPredictedLocation(c *gin.Context) {
	tripID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip ID"})
		return
	}

	trip, err := h.tripRepo.GetByID(uint(tripID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}

	// Check if route data exists
	if trip.RouteStepsJSON == "" || trip.RouteDurationSec == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ルートデータが登録されていません。Trip作成時にルート情報を保存してください。"})
		return
	}

	// Determine target time
	targetTime := time.Now()
	if atParam := c.Query("at"); atParam != "" {
		parsed, err := time.Parse(time.RFC3339, atParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time format. Use RFC3339 (e.g. 2025-01-01T10:00:00+09:00)"})
			return
		}
		targetTime = parsed
	}

	// Calculate elapsed seconds since departure
	elapsedSec := int(targetTime.Sub(trip.DepartureAt).Seconds())

	// Add delay minutes if any
	if trip.DelayMinutes > 0 {
		elapsedSec -= trip.DelayMinutes * 60
	}

	// Predict location
	predicted, err := util.PredictLocationFromSteps(trip.RouteStepsJSON, elapsedSec, trip.RouteDurationSec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "予測位置の計算に失敗しました"})
		return
	}

	if predicted == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ルートステップデータが空です"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trip_id":    trip.ID,
		"trip_status": trip.Status,
		"departure_at": trip.DepartureAt,
		"estimated_arrival": trip.EstimatedArrival,
		"target_time": targetTime,
		"prediction": predicted,
		"origin":      gin.H{"address": trip.OriginAddress, "lat": trip.OriginLat, "lng": trip.OriginLng},
		"destination": gin.H{"address": trip.DestinationAddress, "lat": trip.DestinationLat, "lng": trip.DestinationLng},
	})
}
