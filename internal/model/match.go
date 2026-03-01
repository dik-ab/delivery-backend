package model

import "time"

// Match represents a match between a trip and cargo
type Match struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TripID           uint      `json:"trip_id"`
	Trip             Trip      `json:"trip" gorm:"foreignKey:TripID"`
	ShipperID        uint      `json:"shipper_id"`
	Shipper          User      `json:"shipper" gorm:"foreignKey:ShipperID"`
	CargoWeight      float64   `json:"cargo_weight"`      // kg
	CargoDescription string    `json:"cargo_description"`
	Status           string    `json:"status" gorm:"default:pending"` // pending, approved, rejected, completed
	Message          string    `json:"message"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Match) TableName() string {
	return "matches"
}
