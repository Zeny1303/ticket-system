package utils

import (
	"github.com/gin-gonic/gin"
)

// APIResponse is the standard response envelope for every API response.
// Every endpoint — success or error — returns this shape.
// This consistency makes it easy for frontend clients to handle responses.
//
// Example success: {"success": true, "message": "Login successful", "data": {...}}
// Example error:   {"success": false, "message": "Email already registered", "data": null}
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success sends a successful JSON response.
// Parameters:
//   c       — the Gin context (carries request/response)
//   status  — HTTP status code (e.g., 200, 201)
//   message — human-readable success message
//   data    — the payload to return (struct, slice, map, or nil)
func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error JSON response.
// Parameters:
//   c       — the Gin context
//   status  — HTTP status code (e.g., 400, 401, 404, 500)
//   message — human-readable error message
//
// Note: We don't include raw error details in the response for security.
// Internal error details are logged server-side, not exposed to clients.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
	})
}