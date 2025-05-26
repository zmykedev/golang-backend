package models

import (
	"gorm.io/gorm"
)

type Driver struct {
	gorm.Model
	UserID        uint    `json:"user_id" gorm:"not null"`
	User          User    `json:"user" gorm:"foreignKey:UserID"`
	LicenseNumber string  `json:"license_number" gorm:"not null"`
	VehicleType   string  `json:"vehicle_type" gorm:"not null"`
	VehicleModel  string  `json:"vehicle_model" gorm:"not null"`
	VehicleColor  string  `json:"vehicle_color" gorm:"not null"`
	Languages     string  `json:"languages" gorm:"not null"`  // Comma-separated list of languages
	Experience    int     `json:"experience" gorm:"not null"` // Years of experience
	Rating        float32 `json:"rating" gorm:"default:0"`
	Status        string  `json:"status" gorm:"default:'pending'"` // pending, active, suspended
	IsAvailable   bool    `json:"is_available" gorm:"default:true"`
}
