package routes

import (
	"fiber-backend/database"
	"fiber-backend/middleware"
	"fiber-backend/models"

	"github.com/gofiber/fiber/v2"
)

func SetupTouristRoutes(app *fiber.App) {
	tourist := app.Group("/api/tourists")

	// Create tourist profile
	tourist.Post("/", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var tourist models.Tourist
		if err := c.BodyParser(&tourist); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error al procesar los datos",
			})
		}

		// Set the user ID from the authenticated user
		tourist.UserID = userID

		// Create tourist profile
		if err := database.DB.Create(&tourist).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al crear el perfil de turista",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(tourist)
	})

	// Get tourist profile
	tourist.Get("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var tourist models.Tourist
		if err := database.DB.Where("user_id = ?", userID).First(&tourist).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Perfil de turista no encontrado",
			})
		}

		return c.JSON(tourist)
	})

	// Update tourist profile
	tourist.Put("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var tourist models.Tourist
		if err := database.DB.Where("user_id = ?", userID).First(&tourist).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Perfil de turista no encontrado",
			})
		}

		var updateData models.Tourist
		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error al procesar los datos",
			})
		}

		// Update fields
		tourist.Nationality = updateData.Nationality
		tourist.Language = updateData.Language
		tourist.ArrivalDate = updateData.ArrivalDate
		tourist.DepartureDate = updateData.DepartureDate
		tourist.Preferences = updateData.Preferences
		tourist.SpecialNeeds = updateData.SpecialNeeds

		if err := database.DB.Save(&tourist).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al actualizar el perfil",
			})
		}

		return c.JSON(tourist)
	})

	// Add the new route for requesting a driver
	tourist.Post("/request", middleware.Protected(), RequestDriver)
}

// RequestDriver handles the tourist's request for a driver
func RequestDriver(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Get the tourist profile
	var tourist models.Tourist
	if err := database.DB.Where("user_id = ?", userID).First(&tourist).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Perfil de turista no encontrado",
			"code":  "TOURIST_NOT_FOUND",
		})
	}

	// Parse request data
	var requestData struct {
		PickupLocation  string `json:"pickup_location"`
		DropoffLocation string `json:"dropoff_location"`
		DateTime        string `json:"date_time"`
		Notes           string `json:"notes"`
	}

	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error al procesar los datos",
			"code":  "INVALID_REQUEST_DATA",
		})
	}

	// Create the request
	request := models.TouristRequest{
		TouristID:       tourist.ID,
		PickupLocation:  requestData.PickupLocation,
		DropoffLocation: requestData.DropoffLocation,
		DateTime:        requestData.DateTime,
		Notes:           requestData.Notes,
		Status:          "pending",
	}

	if err := database.DB.Create(&request).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al crear la solicitud",
			"code":  "DATABASE_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Solicitud enviada exitosamente",
		"request": request,
	})
}
