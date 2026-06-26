package repository

import (
	"github.com/ticket-system/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user database operations.
// Using an interface here means:
//   1. The service layer depends on the interface, not the concrete struct
//   2. In tests, you can provide a mock implementation without a real DB
//   3. This is the Dependency Inversion Principle from SOLID
//
// For this assignment, the concrete struct below implements this interface.
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
}

// userRepository is the concrete implementation.
// It is unexported (lowercase) — external packages access it only via
// the UserRepository interface, not directly. This enforces encapsulation.
type userRepository struct {
	// db is the GORM database instance injected via the constructor.
	// We store it as a field so every method can use it without
	// accessing the global database.DB variable.
	// This makes the repository independently testable.
	db *gorm.DB
}

// NewUserRepository is the constructor for userRepository.
// This is Go's standard pattern for constructors — a function that
// returns the interface type. Callers get the interface, not the struct.
//
// In Python/Django, this would be a class instantiation: UserRepository(db)
// In Node/Express, this would be: module.exports = (db) => ({ create, findByEmail })
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database.
// The user.Password must already be hashed before calling this — the
// repository doesn't know about or apply hashing. That's the service's job.
//
// GORM's db.Create() executes: INSERT INTO users (email, password, ...) VALUES (...)
// It also populates user.ID with the generated primary key after insertion.
func (r *userRepository) Create(user *models.User) error {
	result := r.db.Create(user)
	return result.Error
}

// FindByEmail retrieves a user by their email address.
// Used during login to check if the email exists and get the stored hash.
//
// GORM's db.Where().First() executes:
//   SELECT * FROM users WHERE email = ? LIMIT 1
//
// If no user is found, GORM returns gorm.ErrRecordNotFound.
// We return the error as-is — the service layer checks for this specific error
// to distinguish "not found" from a real database error.
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindByID retrieves a user by their primary key ID.
// Used when we need to return user details alongside a token.
//
// GORM's db.First() with a uint argument executes:
//   SELECT * FROM users WHERE id = ? LIMIT 1
func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}