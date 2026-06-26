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

var DB *gorm.DB

func Connect(cfg *config.Config) {
	var err error

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
