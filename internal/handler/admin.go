package handler

import (
	"net/http"
	"strconv"

	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	userRepo  *repository.UserRepository
	tripRepo  *repository.TripRepository
	matchRepo *repository.MatchRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userRepo *repository.UserRepository, tripRepo *repository.TripRepository, matchRepo *repository.MatchRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:  userRepo,
		tripRepo:  tripRepo,
		matchRepo: matchRepo,
	}
}

// StatsResponse represents dashboard statistics
type StatsResponse struct {
	TotalUsers      int `json:"total_users"`
	TotalTrips      int `json:"total_trips"`
	TotalMatches    int `json:"total_matches"`
	ActiveTrips     int `json:"active_trips"`
	CompletedTrips  int `json:"completed_trips"`
	PendingMatches  int `json:"pending_matches"`
	ApprovedMatches int `json:"approved_matches"`
}

// GetStats godoc
// @Summary Get dashboard statistics
// @Description Get platform statistics (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} StatsResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	users, _ := h.userRepo.GetAll()
	trips, _ := h.tripRepo.GetAll()
	matches, _ := h.matchRepo.GetAll()

	activeCount := 0
	completedCount := 0
	for _, trip := range trips {
		if trip.Status == "in_transit" || trip.Status == "matched" {
			activeCount++
		} else if trip.Status == "completed" {
			completedCount++
		}
	}

	pendingCount := 0
	approvedCount := 0
	for _, match := range matches {
		if match.Status == "pending" {
			pendingCount++
		} else if match.Status == "approved" {
			approvedCount++
		}
	}

	stats := StatsResponse{
		TotalUsers:      len(users),
		TotalTrips:      len(trips),
		TotalMatches:    len(matches),
		ActiveTrips:     activeCount,
		CompletedTrips:  completedCount,
		PendingMatches:  pendingCount,
		ApprovedMatches: approvedCount,
	}

	c.JSON(http.StatusOK, stats)
}

// GetUsers godoc
// @Summary Get all users
// @Description Get a list of all platform users (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {array} model.User
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetTrips godoc
// @Summary Get all trips
// @Description Get a list of all trips (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {array} model.Trip
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/trips [get]
func (h *AdminHandler) GetTrips(c *gin.Context) {
	trips, err := h.tripRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch trips"})
		return
	}
	c.JSON(http.StatusOK, trips)
}

// GetMatches godoc
// @Summary Get all matches
// @Description Get a list of all matches (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {array} model.Match
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/matches [get]
func (h *AdminHandler) GetMatches(c *gin.Context) {
	matches, err := h.matchRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}
	c.JSON(http.StatusOK, matches)
}

// UpdateUserRoleRequest represents a user role update request
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=driver shipper admin"`
}

// UpdateUserRole godoc
// @Summary Update user role
// @Description Change user role (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UpdateUserRoleRequest true "Role update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.Role = req.Role
	if err := h.userRepo.Update(uint(id), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user role updated successfully"})
}
