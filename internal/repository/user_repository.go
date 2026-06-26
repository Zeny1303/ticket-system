package repository

import (
	"github.com/Zeny1303/ticket-system/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user database operations.
// The service layer depends on this interface, not the concrete struct,
// enabling mock-based testing without a real database.
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
}

// userRepository is the unexported concrete implementation.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository constructs a userRepository.
// Returns the UserRepository interface — callers never see the concrete type.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database.
// user.Password must already be hashed before calling this.
// GORM: INSERT INTO users (email, password, ...) VALUES (...)
// Populates user.ID with the generated primary key after insertion.
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail retrieves a user by email address.
// Returns gorm.ErrRecordNotFound if no user exists with that email.
// GORM: SELECT * FROM users WHERE email = ? LIMIT 1
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindByID retrieves a user by primary key.
// GORM: SELECT * FROM users WHERE id = ? LIMIT 1
func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
