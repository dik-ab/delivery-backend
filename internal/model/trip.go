package model

import "time"

// Trip represents a driver's trip (delivery route)
type Trip struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	DriverID           uint      `json:"driver_id"`
	Driver             User      `json:"driver" gorm:"foreignKey:DriverID" binding:"-"`
	OriginAddress      string    `json:"origin_address" binding:"required"`
	OriginLat          float64   `json:"origin_lat"`
	OriginLng          float64   `json:"origin_lng"`
	DestinationAddress string    `json:"destination_address" binding:"required"`
	DestinationLat     float64   `json:"destination_lat"`
	DestinationLng     float64   `json:"destination_lng"`
	DepartureAt        time.Time `json:"departure_at"`
	EstimatedArrival   *time.Time `json:"estimated_arrival"`
	VehicleType        string    `json:"vehicle_type"`
	AvailableWeight    float64   `json:"available_weight"` // kg
	Price              int       `json:"price"`            // 円
	Status             string    `json:"status" gorm:"default:open"` // open, matched, in_transit, completed, cancelled
	Note               string    `json:"note"`
	DelayMinutes       int       `json:"delay_minutes" gorm:"default:0"` // 渋滞遅延（分）
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Trip) TableName() string {
	return "trips"
}
