package routes

import (
	"fiber-backend/database"
	"fiber-backend/middleware"
	"fiber-backend/models"

	"log"

	"github.com/gofiber/fiber/v2"
)

func SetupDriverRoutes(app *fiber.App) {
	driver := app.Group("/api/drivers")

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

		// Log parsed struct
		log.Printf("Datos recibidos (antes de setear userID): %+v\n", driver)

		// Set the user ID from the authenticated user
		driver.UserID = userID

		// Log after setting userID
		log.Printf("Datos a guardar (después de setear userID): %+v\n", driver)

		// Create driver profile
		if err := database.DB.Create(&driver).Error; err != nil {
			log.Println("Error al crear el perfil de chofer:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al crear el perfil de chofer",
			})
		}

		// Log after saving
		log.Printf("Chofer guardado en la base de datos: %+v\n", driver)

		return c.Status(fiber.StatusCreated).JSON(driver)
	})

	// Get driver profile
	driver.Get("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var driver models.Driver
		if err := database.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Perfil de chofer no encontrado",
			})
		}

		return c.JSON(driver)
	})

	// Update driver profile
	driver.Put("/me", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var driver models.Driver
		if err := database.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Perfil de chofer no encontrado",
			})
		}

		var updateData models.Driver
		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error al procesar los datos",
			})
		}

		// Update fields
		driver.LicenseNumber = updateData.LicenseNumber
		driver.VehicleType = updateData.VehicleType
		driver.VehicleModel = updateData.VehicleModel
		driver.VehicleColor = updateData.VehicleColor
		driver.Languages = updateData.Languages
		driver.Experience = updateData.Experience
		driver.IsAvailable = updateData.IsAvailable

		if err := database.DB.Save(&driver).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al actualizar el perfil",
			})
		}

		return c.JSON(driver)
	})

	// Get all available drivers
	driver.Get("/available", func(c *fiber.Ctx) error {
		var drivers []models.Driver
		if err := database.DB.Where("is_available = ? AND status = ?", true, "active").Find(&drivers).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al obtener los choferes disponibles",
			})
		}

		return c.JSON(drivers)
	})

	// Update driver availability
	driver.Patch("/me/availability", middleware.Protected(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)

		var driver models.Driver
		if err := database.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Perfil de chofer no encontrado",
			})
		}

		var updateData struct {
			IsAvailable bool `json:"is_available"`
		}

		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error al procesar los datos",
			})
		}

		driver.IsAvailable = updateData.IsAvailable

		if err := database.DB.Save(&driver).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al actualizar la disponibilidad",
			})
		}

		return c.JSON(driver)
	})
}
