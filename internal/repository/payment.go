package repository

import (
	"errors"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// PaymentRepository handles payment database operations
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment record
func (r *PaymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// GetByID retrieves a payment by ID
func (r *PaymentRepository) GetByID(id uint) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Preload("Match").Preload("Payer").First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

// GetByMatchID retrieves all payments for a match
func (r *PaymentRepository) GetByMatchID(matchID uint) ([]model.Payment, error) {
	var payments []model.Payment
	if err := r.db.Where("match_id = ?", matchID).Preload("Payer").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// GetByPayerID retrieves all payments by a user
func (r *PaymentRepository) GetByPayerID(payerID uint) ([]model.Payment, error) {
	var payments []model.Payment
	if err := r.db.Where("payer_id = ?", payerID).Preload("Match").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// Update updates a payment record
func (r *PaymentRepository) Update(id uint, payment *model.Payment) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", id).Updates(payment).Error
}

// GetByStripePaymentID retrieves a payment by Stripe Payment Intent ID
func (r *PaymentRepository) GetByStripePaymentID(stripeID string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("stripe_payment_id = ?", stripeID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}
