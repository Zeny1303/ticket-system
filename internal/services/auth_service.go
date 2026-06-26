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

type AuthService interface {
	Register(req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(req *models.LoginRequest) (*models.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrDatabase
	}
	if existingUser != nil {
		return nil, apperrors.ErrEmailTaken
	}

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

	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiryHours)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

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
