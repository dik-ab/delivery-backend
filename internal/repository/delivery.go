package repository

import (
	"errors"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// DeliveryRepository handles delivery database operations
type DeliveryRepository struct {
	db *gorm.DB
}

// NewDeliveryRepository creates a new delivery repository
func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

// GetAll retrieves all deliveries
func (r *DeliveryRepository) GetAll() ([]model.Delivery, error) {
	var deliveries []model.Delivery
	if err := r.db.Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}

// GetByID retrieves a delivery by ID
func (r *DeliveryRepository) GetByID(id uint) (*model.Delivery, error) {
	var delivery model.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("delivery not found")
		}
		return nil, err
	}
	return &delivery, nil
}

// Create creates a new delivery
func (r *DeliveryRepository) Create(delivery *model.Delivery) error {
	return r.db.Create(delivery).Error
}

// Update updates an existing delivery
func (r *DeliveryRepository) Update(id uint, delivery *model.Delivery) error {
	if err := r.db.Model(&model.Delivery{}).Where("id = ?", id).Updates(delivery).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a delivery by ID
func (r *DeliveryRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Delivery{}, id).Error; err != nil {
		return err
	}
	return nil
}
