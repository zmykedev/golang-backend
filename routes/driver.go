package routes

import (
	"fiber-backend/middleware"
	"fiber-backend/models"
	"fiber-backend/services"
	"log"

	"github.com/gofiber/fiber/v2"
)

func SetupDriverRoutes(app *fiber.App, driverService *services.DriverService) {
	driver := app.Group("/api/drivers")

	// Get all drivers
	driver.Get("/", func(c *fiber.Ctx) error {
		drivers, err := driverService.GetAllDrivers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al obtener los choferes",
			})
		}
		return c.JSON(drivers)
	})

	// Create driver profile
	driver.Post("/", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		// Log raw request body
		log.Println("Body crudo:", string(c.Body()))

		var driver models.Driver
		if err := c.BodyParser(&driver); err != nil {
			log.Println("Error al parsear el body:", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error al procesar los datos",
			})
		}

		// Set the user ID from the authenticated user
		driver.UserID = userID

		// Create driver profile using service
		if err := driverService.CreateDriver(&driver); err != nil {
			log.Println("Error al crear el perfil de chofer:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al crear el perfil de chofer",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(driver)
	})

	// Get driver profile
	driver.Get("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		driver, err := driverService.GetDriverByUserID(userID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Perfil de chofer no encontrado",
			})
		}

		return c.JSON(driver)
	})

	// Get all available drivers
	driver.Get("/available", func(c *fiber.Ctx) error {
		drivers, err := driverService.GetAvailableDrivers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al obtener los choferes disponibles",
			})
		}

		return c.JSON(drivers)
	})

	// Update driver availability
	driver.Patch("/me/availability", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var updateData struct {
			IsAvailable bool `json:"is_available"`
		}

		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error al procesar los datos",
			})
		}

		if err := driverService.UpdateDriverAvailability(userID, updateData.IsAvailable); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al actualizar la disponibilidad",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Disponibilidad actualizada exitosamente",
		})
	})
}
