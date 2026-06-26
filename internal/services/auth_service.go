package services

import (
	"errors"

	"github.com/Zeny1303/ticket-system/internal/config"
	"github.com/Zeny1303/ticket-system/internal/models"
	"github.com/Zeny1303/ticket-system/internal/repository"
	"github.com/Zeny1303/ticket-system/pkg/apperrors"
	"github.com/Zeny1303/ticket-system/pkg/utils"
	"gorm.io/gorm"
)

// AuthService defines the interface for authentication business logic.
type AuthService interface {
	Register(req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(req *models.LoginRequest) (*models.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewAuthService constructs an authService with injected dependencies.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register creates a new user account.
// Business rules:
//  1. Email must not already be registered
//  2. Password is hashed before storage
//  3. A JWT token is returned immediately (auto-login on register)
func (s *authService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Rule 1: check for duplicate email.
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrDatabase
	}
	if existingUser != nil {
		// Issue #12 fix: return sentinel error instead of raw string.
		return nil, apperrors.ErrEmailTaken
	}

	// Rule 2: hash the password before storage.
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Rule 3: generate JWT for auto-login after registration.
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login authenticates an existing user and returns a JWT.
// Business rules:
//  1. Email must exist
//  2. Password must match stored hash
//  3. Returns a JWT on success
//
// Security: same error message for "not found" and "wrong password"
// to prevent user enumeration.
func (s *authService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, apperrors.ErrDatabase
	}

	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}
