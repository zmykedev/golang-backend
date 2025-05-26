package main

import (
	"fiber-backend/config"
	"fiber-backend/database"
	"fiber-backend/routes"
	"fiber-backend/services"
	"fiber-backend/utils"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	utils.LogInfo("Starting application")

	// Load environment variables based on environment
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			utils.LogInfo("No .env file found")
		}
	}

	// Connect to database
	utils.LogInfo("Connecting to database")
	database.Connect()

	// Initialize OAuth configuration
	utils.LogInfo("Initializing OAuth configuration")
	config.InitOAuth()

	// Create a new Fiber instance with custom config
	app := fiber.New(fiber.Config{
		AppName:      "Fiber Auth API",
		ErrorHandler: utils.ErrorHandler,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://tourist-golang.netlify.app/",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH",
	}))

	// Initialize services
	authService := services.NewAuthService(database.DB)

	// Setup routes
	utils.LogInfo("Setting up routes")
	routes.SetupAuthRoutes(app, authService)
	routes.SetupTouristRoutes(app)
	routes.SetupDriverRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	utils.LogInfo("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
