package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evabharat/ticket-system/internal/config"
	"github.com/evabharat/ticket-system/internal/database"
	"github.com/evabharat/ticket-system/internal/handlers"
	"github.com/evabharat/ticket-system/internal/repository"
	"github.com/evabharat/ticket-system/internal/routes"
	"github.com/evabharat/ticket-system/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// =========================================================
	// Step 1: Load Configuration
	// Read all environment variables and build the Config struct.
	// The app exits here if JWT_SECRET is missing.
	// =========================================================
	cfg := config.Load()
	log.Printf("Configuration loaded. Server will start on port %s", cfg.Port)

	// =========================================================
	// Step 2: Connect to Database
	// Establishes PostgreSQL connection via GORM and runs migrations.
	// The app exits here if the DB is unreachable.
	// =========================================================
	database.Connect(cfg)
	log.Println("Database ready.")

	// =========================================================
	// Step 3: Initialize Repository Layer
	// Repositories are injected with the GORM DB instance.
	// They are the only layer that touches the database directly.
	// =========================================================
	userRepo := repository.NewUserRepository(database.DB)
	ticketRepo := repository.NewTicketRepository(database.DB)

	// =========================================================
	// Step 4: Initialize Service Layer
	// Services receive repositories and config as dependencies.
	// Services contain all business logic.
	// =========================================================
	authService := services.NewAuthService(userRepo, cfg)
	ticketService := services.NewTicketService(ticketRepo)

	// =========================================================
	// Step 5: Initialize Handler Layer
	// Handlers receive services as dependencies.
	// Handlers know about HTTP — they translate between HTTP and service calls.
	// =========================================================
	authHandler := handlers.NewAuthHandler(authService)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	// =========================================================
	// Step 6: Set Up the Gin Router
	// gin.Default() creates a router with Logger and Recovery middleware.
	// Logger — logs every request: method, path, status, latency
	// Recovery — catches any panics and returns 500 instead of crashing
	// =========================================================
	router := gin.Default()

	// Register all routes with their handlers and middleware.
	routes.Setup(router, cfg, authHandler, ticketHandler)

	// =========================================================
	// Step 7: Start the HTTP Server
	// We use net/http.Server directly (instead of router.Run()) because
	// it gives us control over timeouts and graceful shutdown.
	// =========================================================
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
		// Timeouts prevent slow clients from holding connections open indefinitely.
		// This is a production best practice — Gin's router.Run() doesn't set these.
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine so it doesn't block the main goroutine.
	// A goroutine is Go's lightweight concurrent execution unit.
	// Django equivalent: the WSGI server runs in a thread. In Go, we use goroutines.
	// The "go" keyword spawns a new goroutine — it runs concurrently with main().
	go func() {
		log.Printf("Server starting on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: Server failed to start: %v", err)
		}
	}()

	// =========================================================
	// Step 8: Graceful Shutdown
	// Wait for an OS signal (Ctrl+C or SIGTERM from Docker/Railway).
	// When received, give in-flight requests 10 seconds to complete
	// before shutting down. This prevents dropped requests during deployments.
	//
	// Without graceful shutdown, a deployment would kill the process
	// mid-request, causing errors for any user whose request was being processed.
	// =========================================================

	// make(chan os.Signal, 1) creates a buffered channel of size 1.
	// Channels are Go's way of communicating between goroutines.
	// This channel will receive OS signals.
	quit := make(chan os.Signal, 1)

	// signal.Notify tells the OS to send SIGINT (Ctrl+C) and SIGTERM (kill/Docker stop)
	// to our quit channel instead of immediately terminating the process.
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block main() until a signal is received on the quit channel.
	// The arrow <- reads from the channel. This line blocks until a signal arrives.
	<-quit
	log.Println("Shutdown signal received. Shutting down gracefully...")

	// Create a context with a 10-second deadline.
	// If in-flight requests don't complete within 10 seconds, force shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("FATAL: Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}