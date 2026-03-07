package model

import "time"

// Payment represents a payment record for a match
type Payment struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MatchID         uint      `json:"match_id"`
	Match           Match     `json:"match" gorm:"foreignKey:MatchID" binding:"-"`
	PayerID         uint      `json:"payer_id"`
	Payer           User      `json:"payer" gorm:"foreignKey:PayerID" binding:"-"`
	Amount          int       `json:"amount"`                                           // 円
	Currency        string    `json:"currency" gorm:"default:jpy"`
	StripePaymentID string    `json:"stripe_payment_id" gorm:"type:varchar(255)"`       // Stripe Payment Intent ID
	Status          string    `json:"status" gorm:"default:pending"`                    // pending, succeeded, failed, refunded
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Payment) TableName() string {
	return "payments"
}
