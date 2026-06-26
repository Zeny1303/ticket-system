package middleware

import (
	"net/http"
	"strings"

	"github.com/ticket-system/internal/config"
	"github.com/ticket-system/internal/utils"
	"github.com/gin-gonic/gin"
)

// UserIDKey is the key used to store and retrieve the user ID from Gin's context.
// Using a typed constant (instead of a raw string) prevents key collisions
// and makes the code self-documenting.
// Any handler that needs the authenticated user ID uses this exact key.
const UserIDKey = "userID"

// AuthMiddleware returns a Gin middleware function that validates JWT tokens.
// It is a closure — it captures the config and returns the actual handler function.
// This pattern (returning a function) is how Gin middleware is written.
//
// Why return a function instead of being a function directly?
// Because we need to inject the config (for the JWT secret) into the middleware.
// This is Go's equivalent of dependency injection for middleware.
//
// Django equivalent: In DRF, you set permission_classes = [IsAuthenticated]
// Express equivalent: const authMiddleware = (req, res, next) => { ... }
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	// gin.HandlerFunc is a type alias for func(*gin.Context).
	// Every Gin handler and middleware must have this signature.
	return func(c *gin.Context) {
		// Step 1: Extract the Authorization header.
		// The expected format is: "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header is required")
			// c.Abort() stops the middleware chain — the handler will NOT be called.
			// Without Abort(), Gin would continue to the next handler even after
			// we've called c.JSON(). This is a common beginner mistake.
			c.Abort()
			return
		}

		// Step 2: Validate the "Bearer " prefix and extract the token.
		// strings.SplitN splits into at most 2 parts: ["Bearer", "<token>"]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "Authorization header format must be: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			utils.Error(c, http.StatusUnauthorized, "Token cannot be empty")
			c.Abort()
			return
		}

		// Step 3: Validate the JWT token using our utility function.
		// This checks the signature AND expiry.
		claims, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Step 4: Store the authenticated user's ID in Gin's context.
		// c.Set() is a key-value store on the request context.
		// Downstream handlers retrieve this with c.Get(middleware.UserIDKey).
		// This is equivalent to req.user in Express or request.user in Django.
		c.Set(UserIDKey, claims.UserID)

		// Step 5: Call c.Next() to proceed to the actual handler.
		// In Express, this is "next()". In Django, it's "return self.get_response(request)".
		c.Next()
	}
}

// GetUserID is a helper used inside handlers to extract the authenticated user's ID
// from Gin's context. It avoids repeating the type assertion in every handler.
//
// Returns 0 if the user ID is not set (which shouldn't happen on protected routes
// since AuthMiddleware would have aborted the request before reaching the handler).
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	// c.Get returns interface{} — we must type assert to get the actual uint value.
	// The comma-ok pattern: if the assertion fails, id is 0 and ok is false.
	id, ok := userID.(uint)
	if !ok {
		return 0
	}
	return id
}