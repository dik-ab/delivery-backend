package model

import "time"

// User represents a platform user (driver, shipper, or admin)
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"type:varchar(255);uniqueIndex" binding:"required"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name" binding:"required"`
	Role         string    `json:"role" gorm:"default:driver"` // driver, shipper, admin
	Company      string    `json:"company"`
	Phone        string    `json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}
