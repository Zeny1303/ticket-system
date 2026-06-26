package services

import (
	"errors"

	"github.com/ticket-system/internal/config"
	"github.com/ticket-system/internal/models"
	"github.com/ticket-system/internal/repository"
	"github.com/ticket-system/internal/utils"
	"gorm.io/gorm"
)

// AuthService defines the interface for authentication business logic.
// The handler depends on this interface — not the concrete struct.
// This is the Dependency Inversion Principle: high-level modules (handlers)
// depend on abstractions (interfaces), not low-level modules (concrete services).
type AuthService interface {
	Register(req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(req *models.LoginRequest) (*models.AuthResponse, error)
}

// authService is the concrete implementation.
// It holds references to its dependencies — the repository and config.
// These are injected via the constructor, not hardcoded.
type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewAuthService constructs an authService with its dependencies injected.
// This is called in main.go where we wire everything together.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register creates a new user account.
// Business rules enforced here:
//   1. Email must not already be registered
//   2. Password must be hashed before storage
//   3. A JWT token is generated and returned immediately (auto-login on register)
//
// Returns an AuthResponse (token + user) or an error.
func (s *authService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Rule 1: Check if email already exists.
	// We call FindByEmail — if it returns a user (no error), the email is taken.
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// A real database error occurred (not just "not found")
		return nil, errors.New("database error while checking email")
	}
	if existingUser != nil {
		// Email is already registered — return a specific business error.
		// The handler will translate this to HTTP 409 Conflict.
		return nil, errors.New("email already registered")
	}

	// Rule 2: Hash the password before storing.
	// NEVER store plain-text passwords.
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Build the User model to save.
	// We only set Email and Password — ID, CreatedAt, UpdatedAt are set by GORM.
	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Persist the user to the database.
	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Rule 3: Generate a JWT token for the newly registered user.
	// This implements "auto-login on registration" — the client gets a token
	// immediately and doesn't need to make a separate login request.
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login authenticates an existing user.
// Business rules enforced here:
//   1. Email must exist in the database
//   2. Provided password must match the stored hash
//   3. A JWT token is generated and returned
//
// Security note: We return the same error message ("invalid credentials")
// for both "email not found" and "wrong password". This is intentional —
// giving different errors reveals whether an email is registered, which
// aids attackers in user enumeration attacks.
func (s *authService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Step 1: Find the user by email.
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Email not found — but we return a generic message (see security note above)
			return nil, errors.New("invalid credentials")
		}
		return nil, errors.New("database error")
	}

	// Step 2: Verify the password against the stored hash.
	// utils.CheckPassword uses bcrypt.CompareHashAndPassword internally.
	// Returns nil if the password matches, an error if it doesn't.
	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Step 3: Generate a JWT token for the authenticated user.
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}