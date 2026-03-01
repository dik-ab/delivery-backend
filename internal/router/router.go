package router

import (
	"os"

	"github.com/delivery-app/delivery-api/internal/handler"
	"github.com/delivery-app/delivery-api/internal/middleware"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Initialize repositories
	deliveryRepo := repository.NewDeliveryRepository(db)
	userRepo := repository.NewUserRepository(db)
	tripRepo := repository.NewTripRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	trackingRepo := repository.NewTrackingRepository(db)

	// Initialize handlers
	deliveryHandler := handler.NewDeliveryHandler(deliveryRepo)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your_secret_key"
	}
	authHandler := handler.NewAuthHandler(userRepo, jwtSecret)
	tripHandler := handler.NewTripHandler(tripRepo)
	matchHandler := handler.NewMatchHandler(matchRepo, tripRepo)
	trackingHandler := handler.NewTrackingHandler(trackingRepo, tripRepo)
	adminHandler := handler.NewAdminHandler(userRepo, tripRepo, matchRepo)

	// Health check endpoint
	router.GET("/api/v1/health", deliveryHandler.HealthCheck)

	// TODO: Swagger documentation（swag init 実行後に有効化）
	// swaggerfiles "github.com/swaggo/files"
	// ginswagger "github.com/swaggo/gin-swagger"
	// router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	api := router.Group("/api/v1")

	// Auth routes (public)
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Delivery routes (legacy)
	{
		api.GET("/deliveries", deliveryHandler.GetDeliveries)
		api.GET("/deliveries/:id", deliveryHandler.GetDelivery)
		api.POST("/deliveries", deliveryHandler.CreateDelivery)
		api.PUT("/deliveries/:id", deliveryHandler.UpdateDelivery)
		api.DELETE("/deliveries/:id", deliveryHandler.DeleteDelivery)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		// Trip routes
		protected.GET("/trips", tripHandler.GetTrips)
		protected.GET("/trips/:id", tripHandler.GetTrip)
		protected.POST("/trips", middleware.DriverMiddleware(), tripHandler.CreateTrip)
		protected.PUT("/trips/:id", tripHandler.UpdateTrip)
		protected.DELETE("/trips/:id", tripHandler.DeleteTrip)
		protected.POST("/trips/search", tripHandler.SearchTrips)

		// Match routes
		protected.GET("/matches", matchHandler.GetMatches)
		protected.GET("/matches/:id", matchHandler.GetMatch)
		protected.POST("/matches", middleware.ShipperMiddleware(), matchHandler.CreateMatch)
		protected.PUT("/matches/:id/approve", middleware.DriverMiddleware(), matchHandler.ApproveMatch)
		protected.PUT("/matches/:id/reject", middleware.DriverMiddleware(), matchHandler.RejectMatch)
		protected.PUT("/matches/:id/complete", matchHandler.CompleteMatch)

		// Tracking routes
		protected.POST("/tracking", trackingHandler.RecordLocation)
		protected.GET("/tracking/:trip_id", trackingHandler.GetTrackingHistory)
		protected.GET("/tracking/:trip_id/latest", trackingHandler.GetLatestLocation)

		// Admin routes
		admin := protected.Group("")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/admin/stats", adminHandler.GetStats)
			admin.GET("/admin/users", adminHandler.GetUsers)
			admin.GET("/admin/trips", adminHandler.GetTrips)
			admin.GET("/admin/matches", adminHandler.GetMatches)
			admin.PUT("/admin/users/:id/role", adminHandler.UpdateUserRole)
		}
	}

	return router
}
