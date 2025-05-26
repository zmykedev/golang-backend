package routes

import (
	"fiber-backend/config"
	"fiber-backend/database"
	"fiber-backend/middleware"
	"fiber-backend/models"
	"fiber-backend/services"
	"fiber-backend/utils"
	"fmt"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func SetupAuthRoutes(app *fiber.App, authService *services.AuthService) {
	utils.LogInfo("Setting up authentication routes")

	// Initialize OAuth config
	config.InitOAuth()

	// Auth routes
	auth := app.Group("/auth")

	// Register
	auth.Post("/register", func(c *fiber.Ctx) error {
		utils.LogInfo("Processing registration request")
		var user models.User
		if err := c.BodyParser(&user); err != nil {
			utils.LogError("Failed to parse registration request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			utils.LogError("Failed to hash password: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not hash password",
			})
		}
		user.Password = string(hashedPassword)

		// Create user
		if err := database.DB.Create(&user).Error; err != nil {
			utils.LogError("Failed to create user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not create user",
			})
		}

		utils.LogInfo("User registered successfully: %s", user.Email)
		return c.Status(fiber.StatusCreated).JSON(user)
	})

	// Update user role
	auth.Post("/update-role", middleware.Protected(), func(c *fiber.Ctx) error {
		// Get user ID from context (set by middleware)
		userID := c.Locals("userID").(uint)

		// Parse request body
		var body struct {
			Role string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate role
		if body.Role != "tourist" && body.Role != "driver" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid role. Must be 'tourist' or 'driver'",
			})
		}

		// Update user role in database
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		user.Role = body.Role
		if err := database.DB.Save(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user role",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Role updated successfully",
			"user": fiber.Map{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Name,
				"role":  user.Role,
			},
		})
	})

	// Login
	auth.Post("/login", func(c *fiber.Ctx) error {
		utils.LogInfo("Processing login request")
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&input); err != nil {
			utils.LogError("Failed to parse login request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		var user models.User
		if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			utils.LogError("Login failed - user not found: %s", input.Email)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			utils.LogError("Login failed - invalid password for user: %s", input.Email)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Create token
		tokenString, err := utils.GenerateToken(user)
		if err != nil {
			utils.LogError("Failed to generate token for user: %s", user.Email)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not generate token",
			})
		}

		utils.LogInfo("User logged in successfully: %s", user.Email)
		return c.JSON(fiber.Map{
			"token": tokenString,
			"user":  user,
		})
	})

	// Google OAuth routes
	auth.Get("/google", func(c *fiber.Ctx) error {
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")
		url := fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile",
			clientID,
			redirectURI,
		)
		return c.Redirect(url)
	})

	auth.Get("/google/callback", func(c *fiber.Ctx) error {
		code := c.Query("code")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Authorization code not provided",
			})
		}

		user, token, err := authService.HandleGoogleAuth(code)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Redirect to frontend with token and user info
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "https://tourist-golang.netlify.app/" // Default frontend URL now uses port 5173
		}
		return c.Redirect(fmt.Sprintf("%s/success?token=%s&name=%s&email=%s",
			frontendURL, token, url.QueryEscape(user.Name), url.QueryEscape(user.Email)))
	})

	// Protected route example
	auth.Get("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Usuario no encontrado",
				"code":  "USER_NOT_FOUND",
			})
		}

		return c.JSON(fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"google_id":  user.GoogleID,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		})
	})
}
