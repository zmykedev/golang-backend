package middleware

import (
	"fiber-backend/utils"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		utils.LogInfo("Processing protected route: %s", c.Path())

		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			utils.LogError("Authorization header missing for route: %s", c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.LogError("Invalid authorization header format for route: %s", c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Get the token
		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				utils.LogError("Invalid token signing method for route: %s", c.Path())
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			utils.LogError("Token validation failed for route: %s, error: %v", c.Path(), err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Set user ID in context
			userID := uint(claims["sub"].(float64))
			c.Locals("userID", userID)
			utils.LogInfo("User %d authenticated successfully for route: %s", userID, c.Path())
			return c.Next()
		}

		utils.LogError("Invalid token claims for route: %s", c.Path())
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}
}
