package routes

import (
	"fiber-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupBookingRoutes(app *fiber.App, db *gorm.DB) {
	bookingGroup := app.Group("/api/bookings")

	// Create a new booking
	bookingGroup.Post("/", func(c *fiber.Ctx) error {
		var booking models.Booking
		if err := c.BodyParser(&booking); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Set the booking time
		booking.BookedAt = time.Now()
		booking.Status = "pending"

		// Create the booking
		if err := db.Create(&booking).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create booking",
			})
		}

		return c.Status(201).JSON(booking)
	})

	// Get all bookings for a tourist
	bookingGroup.Get("/tourist/:id", func(c *fiber.Ctx) error {
		touristID := c.Params("id")
		var bookings []models.Booking

		if err := db.Preload("Driver").Preload("Tourist").Where("tourist_id = ?", touristID).Find(&bookings).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch bookings",
			})
		}

		return c.JSON(bookings)
	})

	// Get all bookings for a driver
	bookingGroup.Get("/driver/:id", func(c *fiber.Ctx) error {
		driverID := c.Params("id")
		var bookings []models.Booking

		if err := db.Preload("Driver").Preload("Tourist").Where("driver_id = ?", driverID).Find(&bookings).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch bookings",
			})
		}

		return c.JSON(bookings)
	})

	// Update booking status
	bookingGroup.Patch("/:id/status", func(c *fiber.Ctx) error {
		bookingID := c.Params("id")
		var updateData struct {
			Status string `json:"status"`
		}

		if err := c.BodyParser(&updateData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := db.Model(&models.Booking{}).Where("id = ?", bookingID).Update("status", updateData.Status).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update booking status",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Booking status updated successfully",
		})
	})

	// Get booking details
	bookingGroup.Get("/:id", func(c *fiber.Ctx) error {
		bookingID := c.Params("id")
		var booking models.Booking

		if err := db.Preload("Driver").Preload("Tourist").First(&booking, bookingID).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Booking not found",
			})
		}

		return c.JSON(booking)
	})
}
