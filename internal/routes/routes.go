package routes

import (
	"net/http"

	"github.com/Zeny1303/ticket-system/internal/config"
	"github.com/Zeny1303/ticket-system/internal/handlers"
	"github.com/Zeny1303/ticket-system/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(
	router *gin.Engine,
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	ticketHandler *handlers.TicketHandler,
) {
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})

	router.GET("/health", handlers.HealthCheck)

	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	tickets := router.Group("/tickets")
	tickets.Use(middleware.AuthMiddleware(cfg))
	{
		tickets.POST("", ticketHandler.CreateTicket)
		tickets.GET("", ticketHandler.GetUserTickets)
		tickets.GET("/:id", ticketHandler.GetTicketByID)
		tickets.PATCH("/:id/status", ticketHandler.UpdateTicketStatus)
	}
}
