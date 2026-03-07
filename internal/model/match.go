package model

import "time"

// Match status constants
const (
	MatchStatusPending   = "pending"
	MatchStatusApproved  = "approved"
	MatchStatusRejected  = "rejected"
	MatchStatusCompleted = "completed"
)

// Match represents a matching request between shipper/transport company and a trip
type Match struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TripID           uint      `json:"trip_id"`
	Trip             Trip      `json:"trip" gorm:"foreignKey:TripID" binding:"-"`
	ShipperID        uint      `json:"shipper_id"`
	Shipper          User      `json:"shipper" gorm:"foreignKey:ShipperID" binding:"-"`
	CargoWeight      float64   `json:"cargo_weight"`                                   // kg
	CargoDescription string    `json:"cargo_description"`
	Status           string    `json:"status" gorm:"default:pending"`                  // pending, approved, rejected, completed
	Message          string    `json:"message"`                                        // リクエスト時のメッセージ
	RejectReason     string    `json:"reject_reason"`                                  // 拒否理由
	RequestType      string    `json:"request_type" gorm:"default:shipper_to_company"` // shipper_to_company, company_to_company
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Match) TableName() string {
	return "matches"
}
