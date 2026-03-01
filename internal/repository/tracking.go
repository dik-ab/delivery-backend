package repository

import (
	"errors"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// TrackingRepository handles tracking database operations
type TrackingRepository struct {
	db *gorm.DB
}

// NewTrackingRepository creates a new tracking repository
func NewTrackingRepository(db *gorm.DB) *TrackingRepository {
	return &TrackingRepository{db: db}
}

// GetByTripID retrieves all tracking records for a trip
func (r *TrackingRepository) GetByTripID(tripID uint) ([]model.Tracking, error) {
	var trackings []model.Tracking
	if err := r.db.Where("trip_id = ?", tripID).Find(&trackings).Error; err != nil {
		return nil, err
	}
	return trackings, nil
}

// GetLatestByTripID retrieves the latest tracking record for a trip
func (r *TrackingRepository) GetLatestByTripID(tripID uint) (*model.Tracking, error) {
	var tracking model.Tracking
	if err := r.db.Where("trip_id = ?", tripID).Order("recorded_at DESC").First(&tracking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tracking not found")
		}
		return nil, err
	}
	return &tracking, nil
}

// Create creates a new tracking record
func (r *TrackingRepository) Create(tracking *model.Tracking) error {
	return r.db.Create(tracking).Error
}

// Delete deletes tracking records for a trip
func (r *TrackingRepository) Delete(tripID uint) error {
	if err := r.db.Where("trip_id = ?", tripID).Delete(&model.Tracking{}).Error; err != nil {
		return err
	}
	return nil
}
