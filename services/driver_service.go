package services

import (
	"fiber-backend/models"
	"fmt"

	"gorm.io/gorm"
)

type DriverService struct {
	db *gorm.DB
}

func NewDriverService(db *gorm.DB) *DriverService {
	return &DriverService{db: db}
}

// CreateDriver creates a new driver profile
func (s *DriverService) CreateDriver(driver *models.Driver) error {
	driver.Status = "active"
	return s.db.Create(driver).Error
}

// GetDriverByUserID retrieves a driver by user ID
func (s *DriverService) GetDriverByUserID(userID uint) (*models.Driver, error) {
	var driver models.Driver
	err := s.db.Where("user_id = ?", userID).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetAllDrivers retrieves all drivers
func (s *DriverService) GetAllDrivers() ([]models.Driver, error) {
	var drivers []models.Driver
	err := s.db.Find(&drivers).Error
	return drivers, err
}

// GetAvailableDrivers retrieves all available and active drivers
func (s *DriverService) GetAvailableDrivers() ([]models.Driver, error) {
	var drivers []models.Driver
	err := s.db.Preload("User").Where("status = ?", "active").Find(&drivers).Error
	if err != nil {
		return nil, err
	}

	// Format drivers for frontend
	for i := range drivers {
		if drivers[i].User.Name == "" {
			drivers[i].User.Name = "Driver " + fmt.Sprint(drivers[i].ID) // Default name if empty
		}
	}

	return drivers, nil
}

// UpdateDriver updates a driver's profile
func (s *DriverService) UpdateDriver(driver *models.Driver) error {
	return s.db.Save(driver).Error
}

// UpdateDriverAvailability updates a driver's availability status
func (s *DriverService) UpdateDriverAvailability(userID uint, isAvailable bool) error {
	return s.db.Model(&models.Driver{}).
		Where("user_id = ?", userID).
		Update("is_available", isAvailable).Error
}
