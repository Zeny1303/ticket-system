package database

import (
	"log"
	"os"
	"time"

	"github.com/Zeny1303/ticket-system/internal/config"
	"github.com/Zeny1303/ticket-system/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the package-level GORM database instance.
// Initialized once in Connect() and reused across the application.
// GORM's *gorm.DB is safe for concurrent use — it manages an internal
// connection pool automatically.
var DB *gorm.DB

// Connect initializes the PostgreSQL connection using GORM.
// Called once during application startup in main.go.
// Exits immediately if the connection fails (fail fast).
func Connect(cfg *config.Config) {
	var err error

	// Issue #15 fix: use Silent log level in production to avoid leaking
	// SQL queries (which contain user data) to stdout.
	logLevel := logger.Info
	if os.Getenv("APP_ENV") == "production" {
		logLevel = logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	DB, err = gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}

	// Issue #14 fix: configure connection pool to prevent exhaustion.
	// PostgreSQL default max_connections is 100; keep well under that.
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("FATAL: Failed to get sql.DB from GORM: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully.")

	migrate()
}

// migrate runs GORM AutoMigrate for all models.
// Creates tables if they don't exist and adds missing columns.
// Does NOT drop columns or change column types.
func migrate() {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&models.User{},
		&models.Ticket{},
	)
	if err != nil {
		log.Fatalf("FATAL: Database migration failed: %v", err)
	}

	log.Println("Database migrations completed successfully.")
}
