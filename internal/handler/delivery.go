package handler

import (
	"net/http"
	"strconv"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
)

// DeliveryHandler handles delivery-related requests
type DeliveryHandler struct {
	repo *repository.DeliveryRepository
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(repo *repository.DeliveryRepository) *DeliveryHandler {
	return &DeliveryHandler{repo: repo}
}

// GetDeliveries godoc
// @Summary Get all deliveries
// @Description Get a list of all delivery destinations
// @Tags deliveries
// @Accept json
// @Produce json
// @Success 200 {array} model.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries [get]
func (h *DeliveryHandler) GetDeliveries(c *gin.Context) {
	deliveries, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deliveries"})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}

// GetDelivery godoc
// @Summary Get a delivery by ID
// @Description Get a single delivery destination by its ID
// @Tags deliveries
// @Accept json
// @Produce json
// @Param id path int true "Delivery ID"
// @Success 200 {object} model.Delivery
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries/{id} [get]
func (h *DeliveryHandler) GetDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	delivery, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// CreateDelivery godoc
// @Summary Create a new delivery
// @Description Create a new delivery destination
// @Tags deliveries
// @Accept json
// @Produce json
// @Param delivery body model.Delivery true "Delivery object"
// @Success 201 {object} model.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries [post]
func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
	var delivery model.Delivery
	if err := c.ShouldBindJSON(&delivery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(&delivery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create delivery"})
		return
	}

	c.JSON(http.StatusCreated, delivery)
}

// UpdateDelivery godoc
// @Summary Update a delivery
// @Description Update an existing delivery destination
// @Tags deliveries
// @Accept json
// @Produce json
// @Param id path int true "Delivery ID"
// @Param delivery body model.Delivery true "Delivery object"
// @Success 200 {object} model.Delivery
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries/{id} [put]
func (h *DeliveryHandler) UpdateDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	var delivery model.Delivery
	if err := c.ShouldBindJSON(&delivery); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if delivery exists
	_, err = h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	delivery.ID = uint(id)
	if err := h.repo.Update(uint(id), &delivery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery"})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// DeleteDelivery godoc
// @Summary Delete a delivery
// @Description Delete a delivery destination
// @Tags deliveries
// @Accept json
// @Produce json
// @Param id path int true "Delivery ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries/{id} [delete]
func (h *DeliveryHandler) DeleteDelivery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	// Check if delivery exists
	_, err = h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete delivery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery deleted successfully"})
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/health [get]
func (h *DeliveryHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
