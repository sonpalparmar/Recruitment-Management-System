package main

import (
	"log"
	"os"

	"github.com/GolangAssignment/internal/config"
	"github.com/GolangAssignment/internal/models"
	"github.com/GolangAssignment/internal/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Database DSN
	dsn := "host=" + cfg.DBHost +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" port=" + cfg.DBPort +
		" sslmode=disable"

	// Connect to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.Profile{}, &models.Job{}, &models.Application{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate models: %v", err)
	}

	// Set up Gin router
	router := gin.Default()

	// Initialize routes
	routes.SetupRoutes(router, db, cfg)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
