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
	// Step 1: Load configuration.
	// Reads all environment variables and builds the Config struct.
	// Exits immediately if JWT_SECRET is missing.
	cfg := config.Load()
	log.Printf("Configuration loaded. Server will start on port %s", cfg.Port)

	// Step 2: Connect to database.
	// Establishes PostgreSQL connection via GORM and runs AutoMigrate.
	// Exits immediately if the DB is unreachable.
	database.Connect(cfg)
	log.Println("Database ready.")

	// Step 3: Initialize repository layer.
	// Repositories are the only layer that touches the database directly.
	userRepo := repository.NewUserRepository(database.DB)
	ticketRepo := repository.NewTicketRepository(database.DB)

	// Step 4: Initialize service layer.
	// Services receive repositories and config; contain all business logic.
	authService := services.NewAuthService(userRepo, cfg)
	ticketService := services.NewTicketService(ticketRepo)

	// Step 5: Initialize handler layer.
	// Handlers receive services; translate between HTTP and service calls.
	authHandler := handlers.NewAuthHandler(authService)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	// Step 6: Set up Gin router.
	// Issue #24: use gin.New() + explicit middleware for cleaner production setup.
	// gin.Default() is fine for assignments, but gin.New() gives explicit control.
	router := gin.New()
	router.Use(gin.Logger())   // log every request
	router.Use(gin.Recovery()) // recover from panics, return 500

	// Register all routes with handlers and middleware.
	routes.Setup(router, cfg, authHandler, ticketHandler)

	// Step 7: Start HTTP server with explicit timeouts.
	// Using net/http.Server directly (instead of router.Run()) gives us
	// control over timeouts and graceful shutdown.
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it doesn't block the main goroutine.
	go func() {
		log.Printf("Server starting on http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Server failed to start: %v", err)
		}
	}()

	// Step 8: Graceful shutdown.
	// Wait for SIGINT (Ctrl+C) or SIGTERM (Docker/Railway stop signal).
	// Give in-flight requests 10 seconds to complete before forcing shutdown.
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
