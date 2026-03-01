package handler

import (
	"net/http"
	"strconv"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
)

// MatchHandler handles match-related requests
type MatchHandler struct {
	matchRepo *repository.MatchRepository
	tripRepo  *repository.TripRepository
}

// NewMatchHandler creates a new match handler
func NewMatchHandler(matchRepo *repository.MatchRepository, tripRepo *repository.TripRepository) *MatchHandler {
	return &MatchHandler{
		matchRepo: matchRepo,
		tripRepo:  tripRepo,
	}
}

// GetMatches godoc
// @Summary Get all matches for current user
// @Description Get a list of matches
// @Tags matches
// @Accept json
// @Produce json
// @Success 200 {array} model.Match
// @Failure 500 {object} map[string]string
// @Router /api/v1/matches [get]
func (h *MatchHandler) GetMatches(c *gin.Context) {
	matches, err := h.matchRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}
	c.JSON(http.StatusOK, matches)
}

// GetMatch godoc
// @Summary Get a match by ID
// @Description Get a single match by its ID
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Success 200 {object} model.Match
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/matches/{id} [get]
func (h *MatchHandler) GetMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	match, err := h.matchRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, match)
}

// CreateMatchRequest represents a match creation request
type CreateMatchRequest struct {
	TripID              uint    `json:"trip_id" binding:"required"`
	CargoWeight         float64 `json:"cargo_weight" binding:"required"`
	CargoDescription    string  `json:"cargo_description"`
	Message             string  `json:"message"`
}

// CreateMatch godoc
// @Summary Request a match (shipper)
// @Description Create a new match request from shipper to driver
// @Tags matches
// @Accept json
// @Produce json
// @Param request body CreateMatchRequest true "Match request"
// @Success 201 {object} model.Match
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/matches [post]
func (h *MatchHandler) CreateMatch(c *gin.Context) {
	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get shipper ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Verify trip exists
	trip, err := h.tripRepo.GetByID(req.TripID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
		return
	}

	// Check if trip has enough capacity
	if trip.AvailableWeight < req.CargoWeight {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient cargo capacity"})
		return
	}

	match := &model.Match{
		TripID:           req.TripID,
		ShipperID:        userID.(uint),
		CargoWeight:      req.CargoWeight,
		CargoDescription: req.CargoDescription,
		Message:          req.Message,
		Status:           "pending",
	}

	if err := h.matchRepo.Create(match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create match"})
		return
	}

	c.JSON(http.StatusCreated, match)
}

// UpdateMatchStatusRequest represents a match status update request
type UpdateMatchStatusRequest struct {
	Status  string `json:"status" binding:"required,oneof=approved rejected completed"`
	Message string `json:"message"`
}

// ApproveMatch godoc
// @Summary Approve a match
// @Description Approve a match request (driver only)
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Success 200 {object} model.Match
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/matches/{id}/approve [put]
func (h *MatchHandler) ApproveMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	match, err := h.matchRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	match.Status = "approved"
	if err := h.matchRepo.Update(uint(id), match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match"})
		return
	}

	c.JSON(http.StatusOK, match)
}

// RejectMatch godoc
// @Summary Reject a match
// @Description Reject a match request (driver only)
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Success 200 {object} model.Match
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/matches/{id}/reject [put]
func (h *MatchHandler) RejectMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	match, err := h.matchRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	match.Status = "rejected"
	if err := h.matchRepo.Update(uint(id), match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match"})
		return
	}

	c.JSON(http.StatusOK, match)
}

// CompleteMatch godoc
// @Summary Complete a match
// @Description Mark a match as completed
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "Match ID"
// @Success 200 {object} model.Match
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/matches/{id}/complete [put]
func (h *MatchHandler) CompleteMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	match, err := h.matchRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	match.Status = "completed"
	if err := h.matchRepo.Update(uint(id), match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match"})
		return
	}

	c.JSON(http.StatusOK, match)
}
