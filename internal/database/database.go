package database

import (
	"log"

	"github.com/ticket-system/internal/config"
	"github.com/ticket-system/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the package-level GORM database instance.
// It is initialized once in Connect() and reused across the entire application.
// GORM's *gorm.DB is safe for concurrent use — it manages an internal
// connection pool automatically. This is unlike Django where each request
// gets its own connection from the pool transparently.
var DB *gorm.DB

// Connect initializes the PostgreSQL connection using GORM.
// It is called once during application startup in main.go.
// If the connection fails, the application exits immediately (fail fast).
//
// Parameters:
//   cfg *config.Config — the loaded application configuration
//
// After Connect() returns, the global DB variable is safe to use
// from any package that imports this package.
func Connect(cfg *config.Config) {
	var err error

	// gorm.Config controls GORM's behavior.
	// logger.Info makes GORM print every SQL query to stdout during development.
	// In production, you would use logger.Silent or a structured logger.
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// postgres.Open() creates a GORM dialector for PostgreSQL.
	// cfg.DSN() returns the connection string we built in config.go.
	// gorm.Open() establishes the connection and returns a *gorm.DB instance.
	DB, err = gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		// log.Fatalf prints the message and calls os.Exit(1).
		// The application cannot run without a database — fatal exit is correct.
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully.")

	// Run auto-migrations after connecting.
	migrate()
}

// migrate runs GORM's AutoMigrate for all models.
// AutoMigrate creates tables if they don't exist, and adds missing columns
// to existing tables. It does NOT delete columns or change column types.
//
// Django equivalent: python manage.py migrate
// Express equivalent: sequelize.sync() or running migration files
//
// For a production system with complex schema changes, you would use
// a proper migration tool like golang-migrate. For this assignment,
// AutoMigrate is sufficient and keeps the setup simple.
func migrate() {
	log.Println("Running database migrations...")

	// The order matters — User must be created before Ticket
	// because Ticket has a foreign key referencing the users table.
	err := DB.AutoMigrate(
		&models.User{},
		&models.Ticket{},
	)
	if err != nil {
		log.Fatalf("FATAL: Database migration failed: %v", err)
	}

	log.Println("Database migrations completed successfully.")
}