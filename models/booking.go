package models

import (
	"time"

	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	TouristID       uint      `json:"tourist_id" gorm:"not null"`
	Tourist         Tourist   `json:"tourist" gorm:"foreignKey:TouristID"`
	DriverID        uint      `json:"driver_id" gorm:"not null"`
	Driver          Driver    `json:"driver" gorm:"foreignKey:DriverID"`
	Status          string    `json:"status" gorm:"default:'pending'"` // pending, confirmed, completed, cancelled
	BookedAt        time.Time `json:"booked_at" gorm:"not null"`
	PickupLocation  string    `json:"pickup_location"`
	DropoffLocation string    `json:"dropoff_location"`
	DateTime        string    `json:"date_time"`
}
