package model

import "time"

// Vehicle represents a driver's vehicle
type Vehicle struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	Type        string    `json:"type" binding:"required"` // 軽トラ, 2t, 4t, 10t, 大型
	MaxWeight   float64   `json:"max_weight"`              // kg
	PlateNumber string    `json:"plate_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Vehicle) TableName() string {
	return "vehicles"
}
