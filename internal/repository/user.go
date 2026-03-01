package repository

import (
	"errors"

	"github.com/delivery-app/delivery-api/internal/model"
	"gorm.io/gorm"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]model.User, error) {
	var users []model.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user
func (r *UserRepository) Update(id uint, user *model.User) error {
	if err := r.db.Model(&model.User{}).Where("id = ?", id).Updates(user).Error; err != nil {
		return err
	}
	return nil
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}
