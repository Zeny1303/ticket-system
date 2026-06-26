package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zeny1303/ticket-system/internal/config"
	"github.com/Zeny1303/ticket-system/internal/database"
	"github.com/Zeny1303/ticket-system/internal/handlers"
	"github.com/Zeny1303/ticket-system/internal/repository"
	"github.com/Zeny1303/ticket-system/internal/routes"
	"github.com/Zeny1303/ticket-system/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Printf("Configuration loaded. Server will start on port %s", cfg.Port)

	database.Connect(cfg)
	log.Println("Database ready.")

	userRepo := repository.NewUserRepository(database.DB)
	ticketRepo := repository.NewTicketRepository(database.DB)

	authService := services.NewAuthService(userRepo, cfg)
	ticketService := services.NewTicketService(ticketRepo)

	authHandler := handlers.NewAuthHandler(authService)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	routes.Setup(router, cfg, authHandler, ticketHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received. Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
