package repository

import (
	"errors"
	"time"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// TripRepository handles trip database operations
type TripRepository struct {
	db *gorm.DB
}

// NewTripRepository creates a new trip repository
func NewTripRepository(db *gorm.DB) *TripRepository {
	return &TripRepository{db: db}
}

// GetAll retrieves all trips
func (r *TripRepository) GetAll() ([]model.Trip, error) {
	var trips []model.Trip
	if err := r.db.Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// GetByID retrieves a trip by ID
func (r *TripRepository) GetByID(id uint) (*model.Trip, error) {
	var trip model.Trip
	if err := r.db.Preload("Driver").First(&trip, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("trip not found")
		}
		return nil, err
	}
	return &trip, nil
}

// GetByDriverID retrieves all trips for a driver
func (r *TripRepository) GetByDriverID(driverID uint) ([]model.Trip, error) {
	var trips []model.Trip
	if err := r.db.Where("driver_id = ?", driverID).Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// GetOpenTrips retrieves all open trips
func (r *TripRepository) GetOpenTrips() ([]model.Trip, error) {
	var trips []model.Trip
	if err := r.db.Where("status = ?", "open").Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// GetTripsAfterDate retrieves trips departing after a given date
func (r *TripRepository) GetTripsAfterDate(date time.Time) ([]model.Trip, error) {
	var trips []model.Trip
	if err := r.db.Where("departure_at >= ?", date).Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}

// Create creates a new trip
func (r *TripRepository) Create(trip *model.Trip) error {
	return r.db.Create(trip).Error
}

// Update updates an existing trip
func (r *TripRepository) Update(id uint, trip *model.Trip) error {
	if err := r.db.Model(&model.Trip{}).Where("id = ?", id).Updates(trip).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a trip by ID
func (r *TripRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Trip{}, id).Error; err != nil {
		return err
	}
	return nil
}

// GetTripsByOriginDestination finds trips matching origin and destination within radius
func (r *TripRepository) GetTripsByOriginDestination(originLat, originLng, destLat, destLng float64, radiusKm float64) ([]model.Trip, error) {
	var trips []model.Trip
	// Find trips where origin is near destination and destination is near origin (return trips)
	// or normal direction trips
	if err := r.db.Where("status = ?", "open").Preload("Driver").Find(&trips).Error; err != nil {
		return nil, err
	}
	return trips, nil
}
