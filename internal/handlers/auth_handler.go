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

// AuthHandler holds the dependencies needed by auth HTTP handlers.
type AuthHandler struct {
	authService services.AuthService
	validate    *validator.Validate
}

// NewAuthHandler constructs an AuthHandler.
// The validator instance is created once and reused for all requests.
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// Register handles POST /auth/register
//
// Flow:
//  1. Parse JSON body into RegisterRequest
//  2. Validate struct fields
//  3. Call auth service
//  4. Return 201 Created with token and user, or appropriate error
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
		// Issue #12 fix: use errors.Is() against sentinel errors instead of string comparison.
		if errors.Is(err, apperrors.ErrEmailTaken) {
			utils.Error(c, http.StatusConflict, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "User registered successfully", response)
}

// Login handles POST /auth/login
//
// Flow:
//  1. Parse JSON body
//  2. Validate fields
//  3. Call auth service
//  4. Return 200 OK with token and user, or 401 Unauthorized
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
		// Login failures are always 401. Same message for "not found" and "wrong password"
		// to prevent user enumeration attacks.
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Login successful", response)
}
