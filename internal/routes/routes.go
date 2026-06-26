package routes

import (
	"github.com/ticket-system/internal/config"
	"github.com/ticket-system/internal/handlers"
	"github.com/ticket-system/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Setup registers all routes on the Gin engine.
// It receives all handlers and config as parameters — pure dependency injection.
// No global state is accessed here.
//
// Route structure:
//   GET  /health                   — public, no auth
//   POST /auth/register            — public, no auth
//   POST /auth/login               — public, no auth
//   POST /tickets                  — protected, requires Bearer token
//   GET  /tickets                  — protected
//   GET  /tickets/:id              — protected
//   PATCH /tickets/:id/status      — protected
func Setup(
	router *gin.Engine,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	ticketHandler *handlers.TicketHandler,
) {
	// GET /health — The assignment requires this to be publicly accessible.
	// No auth, no middleware. Any monitoring tool can hit this.
	router.GET("/health", handlers.HealthCheck)

	// Auth routes group — no authentication middleware on these.
	// These are the endpoints used to OBTAIN a token, so they can't require one.
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Ticket routes group — ALL routes here require a valid JWT.
	// middleware.AuthMiddleware(cfg) runs before every handler in this group.
	// If the token is missing or invalid, the middleware aborts with 401
	// and the handler is never called.
	//
	// Gin's Use() applies middleware to the group.
	// The braces {} are just a Go formatting convention for clarity —
	// they don't create a new scope for Gin, only for the Go compiler.
	tickets := router.Group("/tickets")
	tickets.Use(middleware.AuthMiddleware(cfg))
	{
		tickets.POST("", ticketHandler.CreateTicket)
		tickets.GET("", ticketHandler.GetUserTickets)
		tickets.GET("/:id", ticketHandler.GetTicketByID)
		tickets.PATCH("/:id/status", ticketHandler.UpdateTicketStatus)
	}
}