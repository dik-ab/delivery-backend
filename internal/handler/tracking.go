package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
)

// TrackingHandler handles tracking-related requests
type TrackingHandler struct {
	trackingRepo *repository.TrackingRepository
	tripRepo     *repository.TripRepository
}

// NewTrackingHandler creates a new tracking handler
func NewTrackingHandler(trackingRepo *repository.TrackingRepository, tripRepo *repository.TripRepository) *TrackingHandler {
	return &TrackingHandler{
		trackingRepo: trackingRepo,
		tripRepo:     tripRepo,
	}
}

// RecordLocationRequest represents a location tracking request
type RecordLocationRequest struct {
	TripID uint    `json:"trip_id" binding:"required"`
	Lat    float64 `json:"lat" binding:"required"`
	Lng    float64 `json:"lng" binding:"required"`
}

// RecordLocation godoc
// @Summary Record a location update
// @Description Record a new location tracking point for a trip
// @Tags tracking
// @Accept json
// @Produce json
// @Param request body RecordLocationRequest true "Location record"
// @Success 201 {object} model.Tracking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tracking [post]
func (h *TrackingHandler) RecordLocation(c *gin.Context) {
	var req RecordLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify trip exists
	_, err := h.tripRepo.GetByID(req.TripID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}

	tracking := &model.Tracking{
		TripID:     req.TripID,
		Lat:        req.Lat,
		Lng:        req.Lng,
		RecordedAt: time.Now(),
	}

	if err := h.trackingRepo.Create(tracking); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record location"})
		return
	}

	c.JSON(http.StatusCreated, tracking)
}

// GetTrackingHistory godoc
// @Summary Get tracking history for a trip
// @Description Get all location tracking records for a specific trip
// @Tags tracking
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {array} model.Tracking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tracking/{trip_id} [get]
func (h *TrackingHandler) GetTrackingHistory(c *gin.Context) {
	tripID, err := strconv.ParseUint(c.Param("trip_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip ID"})
		return
	}

	// Verify trip exists
	_, err = h.tripRepo.GetByID(uint(tripID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}

	trackings, err := h.trackingRepo.GetByTripID(uint(tripID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tracking history"})
		return
	}

	c.JSON(http.StatusOK, trackings)
}

// GetLatestLocation godoc
// @Summary Get latest location for a trip
// @Description Get the most recent location tracking record for a trip
// @Tags tracking
// @Accept json
// @Produce json
// @Param trip_id path int true "Trip ID"
// @Success 200 {object} model.Tracking
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tracking/{trip_id}/latest [get]
func (h *TrackingHandler) GetLatestLocation(c *gin.Context) {
	tripID, err := strconv.ParseUint(c.Param("trip_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip ID"})
		return
	}

	// Verify trip exists
	_, err = h.tripRepo.GetByID(uint(tripID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}

	tracking, err := h.trackingRepo.GetLatestByTripID(uint(tripID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tracking)
}
