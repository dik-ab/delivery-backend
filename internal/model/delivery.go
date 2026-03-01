package model

import "time"

// Delivery represents a delivery destination
type Delivery struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required"`                 // 配送先名
	Address   string    `json:"address" binding:"required"`              // 住所
	Lat       float64   `json:"lat"`                                     // 緯度
	Lng       float64   `json:"lng"`                                     // 経度
	Status    string    `json:"status" gorm:"default:pending"`           // pending, in_progress, completed
	Note      string    `json:"note"`                                    // メモ
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Delivery) TableName() string {
	return "deliveries"
}
