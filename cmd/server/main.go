package main

import (
	"fmt"
	"log"
	"os"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/router"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Delivery API
// @version 1.0
// @description A delivery route management API with Google Maps integration
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
// @schemes http https
func main() {
	// Load environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "delivery_user")
	dbPassword := getEnv("DB_PASSWORD", "delivery_pass")
	dbName := getEnv("DB_NAME", "delivery_db")
	port := getEnv("PORT", "8080")

	// Setup database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate the models
	if err := db.AutoMigrate(
		&model.Delivery{},
		&model.User{},
		&model.Vehicle{},
		&model.Trip{},
		&model.Match{},
		&model.Tracking{},
		&model.Payment{},
	); err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	log.Println("Database connected and migrated successfully")

	// Setup router
	r := router.SetupRouter(db)

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
