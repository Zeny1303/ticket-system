package handlers

import (
	"net/http"

	"github.com/ticket-system/internal/models"
	"github.com/ticket-system/internal/services"
	"github.com/ticket-system/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthHandler holds the dependencies needed by auth HTTP handlers.
// It is a struct with methods — this is Go's equivalent of a class with methods.
// In Django: a ViewSet class. In Express: a controller object.
type AuthHandler struct {
	authService services.AuthService
	validate    *validator.Validate
}

// NewAuthHandler constructs an AuthHandler.
// The validator instance is created once here and reused for all requests.
// Creating a new validator per-request would be wasteful.
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// Register handles POST /auth/register
//
// Flow:
//   1. Parse JSON request body into RegisterRequest struct
//   2. Validate the struct fields using validator tags
//   3. Call auth service to handle registration logic
//   4. Return 201 Created with token and user, or appropriate error
//
// Django equivalent: a CreateAPIView with a serializer
// Express equivalent: an async route handler with req.body validation
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// c.ShouldBindJSON() parses the request body as JSON into our struct.
	// It returns an error if the body is malformed (not valid JSON) or
	// if the Content-Type header is wrong.
	// Unlike c.BindJSON(), ShouldBindJSON does NOT write a 400 response automatically
	// — we control the error response ourselves.
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// h.validate.Struct() runs all the `validate:"..."` tag rules on the struct.
	// If any validation fails, it returns a ValidationErrors containing all failures.
	if err := h.validate.Struct(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Call the service layer — the handler doesn't know HOW registration works,
	// only that it needs to happen and what to do with the result.
	response, err := h.authService.Register(&req)
	if err != nil {
		// Map service errors to appropriate HTTP status codes.
		// "email already registered" is a conflict (409), not a server error (500).
		if err.Error() == "email already registered" {
			utils.Error(c, http.StatusConflict, err.Error())
			return
		}
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 201 Created is the correct status for successful resource creation.
	// Using 200 OK for creation is a common mistake to avoid.
	utils.Success(c, http.StatusCreated, "User registered successfully", response)
}

// Login handles POST /auth/login
//
// Flow:
//   1. Parse JSON request body
//   2. Validate fields
//   3. Call auth service
//   4. Return 200 OK with token and user, or 401 Unauthorized
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
		// Login failures are always 401 Unauthorized.
		// We use the same error message regardless of whether the email
		// doesn't exist or the password is wrong — this prevents user enumeration.
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	// 200 OK is correct for login — we're not creating a new resource,
	// we're authenticating and receiving an existing resource (the token).
	utils.Success(c, http.StatusOK, "Login successful", response)
}