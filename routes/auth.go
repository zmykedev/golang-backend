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
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func SetupAuthRoutes(app *fiber.App, authService *services.AuthService) {
	utils.LogInfo("Setting up authentication routes")

	// Initialize OAuth config
	config.InitOAuth()

	// Auth routes
	auth := app.Group("/auth")

	// Register with tourist profile
	auth.Post("/register", func(c *fiber.Ctx) error {
		utils.LogInfo("Processing registration request")
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
			Tourist  struct {
				Nationality   string `json:"nationality"`
				Language      string `json:"language"`
				ArrivalDate   string `json:"arrival_date"`
				DepartureDate string `json:"departure_date"`
				Preferences   string `json:"preferences"`
				SpecialNeeds  string `json:"special_needs"`
			} `json:"tourist"`
			Role string `json:"role"`
		}

		if err := c.BodyParser(&input); err != nil {
			utils.LogError("Failed to parse registration request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		utils.LogInfo("Registering new user - Email: %s, Name: %s", input.Email, input.Name)
		utils.LogInfo("Raw password length: %d", len(input.Password))

		// Start transaction
		tx := database.DB.Begin()

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
		if err != nil {
			tx.Rollback()
			utils.LogError("Failed to hash password: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not hash password",
			})
		}
		utils.LogInfo("Password hashed successfully - Hash length: %d, Hash: %s", len(hashedPassword), string(hashedPassword))

		// Create user
		user := models.User{
			Email:    input.Email,
			Password: string(hashedPassword),
			Name:     input.Name,
			GoogleID: nil,
			Role:     input.Role,
		}

		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			utils.LogError("Failed to create user in database: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not create user",
			})
		}
		utils.LogInfo("User created successfully in database - ID: %d", user.ID)

		// If registering as a tourist, create tourist profile
		if input.Role == "tourist" {
			arrivalDate, err := time.Parse("2006-01-02", input.Tourist.ArrivalDate)
			if err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid arrival date format. Use YYYY-MM-DD",
				})
			}

			departureDate, err := time.Parse("2006-01-02", input.Tourist.DepartureDate)
			if err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid departure date format. Use YYYY-MM-DD",
				})
			}

			tourist := models.Tourist{
				UserID:        user.ID,
				Nationality:   input.Tourist.Nationality,
				Language:      input.Tourist.Language,
				ArrivalDate:   arrivalDate,
				DepartureDate: departureDate,
				Preferences:   input.Tourist.Preferences,
				SpecialNeeds:  input.Tourist.SpecialNeeds,
				Status:        "pending",
			}

			if err := tx.Create(&tourist).Error; err != nil {
				tx.Rollback()
				utils.LogError("Failed to create tourist profile: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Could not create tourist profile",
				})
			}
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			utils.LogError("Failed to commit transaction: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to complete registration",
			})
		}

		// Generate token for immediate login
		token, err := utils.GenerateToken(user)
		if err != nil {
			utils.LogError("Failed to generate token: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not generate token",
			})
		}

		utils.LogInfo("User registered successfully: %s", user.Email)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"user": fiber.Map{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Name,
				"role":  user.Role,
			},
			"token": token,
		})
	})

	// Update user role
	auth.Post("/update-role", middleware.Protected(), func(c *fiber.Ctx) error {
		// Get user ID from context (set by middleware)
		userID := c.Locals("userID").(uint)
		utils.LogInfo("Processing role update request for user ID: %d", userID)

		// Parse request body
		var body struct {
			Role string `json:"role"`
		}
		if err := c.BodyParser(&body); err != nil {
			utils.LogError("Failed to parse role update request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		utils.LogInfo("Requested role update for user %d to role: %s", userID, body.Role)

		// Validate role
		if body.Role != "tourist" && body.Role != "driver" {
			utils.LogError("Invalid role requested: %s", body.Role)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid role. Must be 'tourist' or 'driver'",
			})
		}

		// Start transaction
		tx := database.DB.Begin()

		// Get current user
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			tx.Rollback()
			utils.LogError("Failed to find user %d: %v", userID, err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// Check if user already has a role
		if user.Role != "" {
			utils.LogInfo("User %d already has role: %s", userID, user.Role)
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User already has a role assigned",
				"role":  user.Role,
			})
		}

		previousRole := user.Role
		user.Role = body.Role

		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			utils.LogError("Failed to update role for user %d from %s to %s: %v", userID, previousRole, body.Role, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user role",
			})
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			utils.LogError("Failed to commit role update transaction: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user role",
			})
		}

		utils.LogInfo("Successfully updated role for user %d from %s to %s", userID, previousRole, user.Role)
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

		utils.LogInfo("Login attempt - Email: %s, Password length: %d", input.Email, len(input.Password))

		var user models.User
		if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			utils.LogError("Login failed - User not found in database: %s", input.Email)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		utils.LogInfo("User found in database - ID: %d, Email: %s", user.ID, user.Email)
		utils.LogInfo("Stored password hash: %s", user.Password)
		utils.LogInfo("Input password length: %d, Stored hash length: %d", len(input.Password), len(user.Password))

		// Compare passwords
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
		if err != nil {
			utils.LogError("Password comparison failed - Error: %v", err)
			utils.LogError("Attempted to compare hash '%s' with password of length %d", user.Password, len(input.Password))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		utils.LogInfo("Password verified successfully")

		// Create token
		token, err := utils.GenerateToken(user)
		if err != nil {
			utils.LogError("Failed to generate token for user: %s", user.Email)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Could not generate token",
			})
		}

		utils.LogInfo("Login successful - User: %s, Token generated", user.Email)
		return c.JSON(fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Name,
				"role":  user.Role,
			},
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
			frontendURL = "http://localhost:5173/success" // Default frontend URL now uses port 5173
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
