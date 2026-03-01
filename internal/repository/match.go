package repository

import (
	"errors"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// MatchRepository handles match database operations
type MatchRepository struct {
	db *gorm.DB
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db *gorm.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

// GetAll retrieves all matches
func (r *MatchRepository) GetAll() ([]model.Match, error) {
	var matches []model.Match
	if err := r.db.Preload("Trip").Preload("Shipper").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

// GetByID retrieves a match by ID
func (r *MatchRepository) GetByID(id uint) (*model.Match, error) {
	var match model.Match
	if err := r.db.Preload("Trip").Preload("Shipper").First(&match, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("match not found")
		}
		return nil, err
	}
	return &match, nil
}

// GetByTripID retrieves all matches for a trip
func (r *MatchRepository) GetByTripID(tripID uint) ([]model.Match, error) {
	var matches []model.Match
	if err := r.db.Where("trip_id = ?", tripID).Preload("Trip").Preload("Shipper").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

// GetByShipperID retrieves all matches for a shipper
func (r *MatchRepository) GetByShipperID(shipperID uint) ([]model.Match, error) {
	var matches []model.Match
	if err := r.db.Where("shipper_id = ?", shipperID).Preload("Trip").Preload("Shipper").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

// GetPendingMatches retrieves all pending matches
func (r *MatchRepository) GetPendingMatches() ([]model.Match, error) {
	var matches []model.Match
	if err := r.db.Where("status = ?", "pending").Preload("Trip").Preload("Shipper").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

// Create creates a new match
func (r *MatchRepository) Create(match *model.Match) error {
	return r.db.Create(match).Error
}

// Update updates an existing match
func (r *MatchRepository) Update(id uint, match *model.Match) error {
	if err := r.db.Model(&model.Match{}).Where("id = ?", id).Updates(match).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a match by ID
func (r *MatchRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Match{}, id).Error; err != nil {
		return err
	}
	return nil
}
