package handlers

import (
	"errors"
	"net/http"

	"github.com/Zeny1303/ticket-system/internal/models"
	"github.com/Zeny1303/ticket-system/internal/services"
	"github.com/Zeny1303/ticket-system/pkg/apperrors"
	"github.com/Zeny1303/ticket-system/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService services.AuthService
	validate    *validator.Validate
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmailTaken) {
			utils.Error(c, http.StatusConflict, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "User registered successfully", response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Login successful", response)
}
