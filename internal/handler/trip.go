package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/delivery-app/delivery-api/internal/util"
	"github.com/gin-gonic/gin"
)

// TripHandler handles trip-related requests
type TripHandler struct {
	tripRepo *repository.TripRepository
}

// NewTripHandler creates a new trip handler
func NewTripHandler(tripRepo *repository.TripRepository) *TripHandler {
	return &TripHandler{tripRepo: tripRepo}
}

// GetTrips godoc
// @Summary Get all trips with filters
// @Description Get a list of trips with optional filters
// @Tags trips
// @Accept json
// @Produce json
// @Param status query string false "Trip status (open, matched, in_transit, completed, cancelled)"
// @Param driver_id query int false "Driver ID"
// @Success 200 {array} model.Trip
// @Failure 500 {object} map[string]string
// @Router /api/v1/trips [get]
func (h *TripHandler) GetTrips(c *gin.Context) {
	trips, err := h.tripRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch trips"})
		return
	}
	c.JSON(http.StatusOK, trips)
}

// GetTrip godoc
// @Summary Get a trip by ID
// @Description Get a single trip by its ID
// @Tags trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} model.Trip
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/trips/{id} [get]
func (h *TripHandler) GetTrip(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip ID"})
		return
	}

	trip, err := h.tripRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// CreateTrip godoc
// @Summary Create a new trip
// @Description Create a new trip (driver only)
// @Tags trips
// @Accept json
// @Produce json
// @Param trip body model.Trip true "Trip object"
// @Success 201 {object} model.Trip
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/trips [post]
func (h *TripHandler) CreateTrip(c *gin.Context) {
	var trip model.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get driver ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	trip.DriverID = userID.(uint)
	trip.Status = "open"

	if err := h.tripRepo.Create(&trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create trip"})
		return
	}

	c.JSON(http.StatusCreated, trip)
}

// UpdateTrip godoc
// @Summary Update a trip
// @Description Update an existing trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param trip body model.Trip true "Trip object"
// @Success 200 {object} model.Trip
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/trips/{id} [put]
func (h *TripHandler) UpdateTrip(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip ID"})
		return
	}

	var trip model.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if trip exists
	_, err = h.tripRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	trip.ID = uint(id)
	if err := h.tripRepo.Update(uint(id), &trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update trip"})
		return
	}

	c.JSON(http.StatusOK, trip)
}

// DeleteTrip godoc
// @Summary Delete a trip
// @Description Delete a trip
// @Tags trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/trips/{id} [delete]
func (h *TripHandler) DeleteTrip(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip ID"})
		return
	}

	// Check if trip exists
	_, err = h.tripRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.tripRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete trip"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "trip deleted successfully"})
}

// SearchRequest represents a trip search request
type SearchRequest struct {
	OriginLat  float64 `json:"origin_lat" binding:"required"`
	OriginLng  float64 `json:"origin_lng" binding:"required"`
	DestLat    float64 `json:"dest_lat" binding:"required"`
	DestLng    float64 `json:"dest_lng" binding:"required"`
	RadiusKm   float64 `json:"radius_km"`
	Date       string  `json:"date"`
}

// SearchTrips godoc
// @Summary Search for return trips
// @Description Search for trips with return trip matching (帰り便シェア)
// @Tags trips
// @Accept json
// @Produce json
// @Param request body SearchRequest true "Search request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/trips/search [post]
func (h *TripHandler) SearchTrips(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RadiusKm == 0 {
		req.RadiusKm = 50 // Default 50km radius
	}

	// Get all open trips
	allTrips, err := h.tripRepo.GetOpenTrips()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search trips"})
		return
	}

	var normalMatches []model.Trip
	var returnMatches []model.Trip

	for _, trip := range allTrips {
		// Filter by date if provided
		if req.Date != "" {
			searchDate, err := time.Parse("2006-01-02", req.Date)
			if err == nil {
				tripDate := trip.DepartureAt.Truncate(24 * time.Hour)
				if tripDate != searchDate.Truncate(24*time.Hour) {
					continue
				}
			}
		}

		// Normal direction: trip origin near search origin, trip destination near search destination
		if util.IsWithinRadius(trip.OriginLat, trip.OriginLng, req.OriginLat, req.OriginLng, req.RadiusKm) &&
			util.IsWithinRadius(trip.DestinationLat, trip.DestinationLng, req.DestLat, req.DestLng, req.RadiusKm) {
			normalMatches = append(normalMatches, trip)
			continue
		}

		// Return trip (帰り便): trip origin near search destination, trip destination near search origin
		if util.IsWithinRadius(trip.OriginLat, trip.OriginLng, req.DestLat, req.DestLng, req.RadiusKm) &&
			util.IsWithinRadius(trip.DestinationLat, trip.DestinationLng, req.OriginLat, req.OriginLng, req.RadiusKm) {
			returnMatches = append(returnMatches, trip)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"normal_matches":      normalMatches,
		"return_matches":      returnMatches,
		"total_matches":       len(normalMatches) + len(returnMatches),
		"normal_count":        len(normalMatches),
		"return_count":        len(returnMatches),
	})
}
