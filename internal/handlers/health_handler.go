package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck handles GET /health
//
// This is the simplest handler in the system — it always returns 200 OK.
// Every deployed service needs a health check endpoint because:
//   1. Deployment platforms (Railway, Render, Fly.io) ping /health to know if the app is running
//   2. Load balancers use it to decide which instances to route traffic to
//   3. The assignment explicitly requires it and tests it
//   4. It confirms the server is up without needing authentication
//
// The response format matches exactly what the assignment specifies:
//   {"status": "ok"}
//
// Note: This is a standalone function, not a method on a struct.
// It has no dependencies — no service, no database.
// There's no reason to wrap it in a struct just for consistency.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}