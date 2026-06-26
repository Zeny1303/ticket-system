package routes

import (
	"net/http"

	"github.com/Zeny1303/ticket-system/internal/config"
	"github.com/Zeny1303/ticket-system/internal/handlers"
	"github.com/Zeny1303/ticket-system/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Setup registers all routes on the Gin engine.
// All dependencies are injected — no global state is accessed here.
//
// Route structure:
//
//	GET  /health                   — public
//	POST /auth/register            — public
//	POST /auth/login               — public
//	POST /tickets                  — protected (JWT required)
//	GET  /tickets                  — protected
//	GET  /tickets/:id              — protected
//	PATCH /tickets/:id/status      — protected
func Setup(
	router *gin.Engine,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	ticketHandler *handlers.TicketHandler,
) {
	// Issue #25 fix: limit request body size to 1 MB to prevent OOM attacks.
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MB
		c.Next()
	})

	// GET /health — publicly accessible, no auth required.
	router.GET("/health", handlers.HealthCheck)

	// Auth routes — no authentication middleware (used to obtain tokens).
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Ticket routes — all require a valid JWT Bearer token.
	// AuthMiddleware aborts with 401 if the token is missing or invalid.
	tickets := router.Group("/tickets")
	tickets.Use(middleware.AuthMiddleware(cfg))
	{
		tickets.POST("", ticketHandler.CreateTicket)
		tickets.GET("", ticketHandler.GetUserTickets)
		tickets.GET("/:id", ticketHandler.GetTicketByID)
		tickets.PATCH("/:id/status", ticketHandler.UpdateTicketStatus)
	}
}
