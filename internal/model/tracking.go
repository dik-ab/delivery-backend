package model

import "time"

// Tracking represents a trip's location tracking data
type Tracking struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TripID     uint      `json:"trip_id"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	RecordedAt time.Time `json:"recorded_at"`
}

// TableName specifies the table name for GORM
func (Tracking) TableName() string {
	return "trackings"
}
