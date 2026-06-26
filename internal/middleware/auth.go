package middleware

import (
	"net/http"
	"strings"

	"github.com/Zeny1303/ticket-system/internal/config"
	"github.com/Zeny1303/ticket-system/pkg/utils"
	"github.com/gin-gonic/gin"
)

// UserIDKey is the context key used to store and retrieve the authenticated user ID.
// Using a typed constant prevents key collisions and makes lookups self-documenting.
const UserIDKey = "userID"

// AuthMiddleware returns a Gin middleware function that validates JWT tokens.
// It is a closure — captures cfg and returns the actual handler func.
//
// On every protected request it:
//  1. Extracts the Authorization header
//  2. Validates the "Bearer <token>" format
//  3. Validates the JWT (signature + expiry)
//  4. Stores the user ID in Gin context for downstream handlers
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: require the Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Step 2: validate "Bearer <token>" format.
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header format must be: Bearer <token>",
				"data":    nil,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token cannot be empty",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Step 3: validate the JWT token (signature + expiry).
		claims, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Step 4: store the authenticated user's ID in Gin context.
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}

// GetUserID extracts the authenticated user's ID from Gin context.
// Issue #13 fix: returns (uint, bool) so callers can detect the missing case
// instead of silently receiving 0.
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}
