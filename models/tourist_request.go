package models

import (
	"gorm.io/gorm"
)

type TouristRequest struct {
	gorm.Model
	TouristID       uint    `json:"tourist_id" gorm:"not null"`
	Tourist         Tourist `json:"tourist" gorm:"foreignKey:TouristID"`
	PickupLocation  string  `json:"pickup_location" gorm:"not null"`
	DropoffLocation string  `json:"dropoff_location" gorm:"not null"`
	DateTime        string  `json:"date_time" gorm:"not null"`
	Notes           string  `json:"notes"`
	Status          string  `json:"status" gorm:"type:varchar(20);default:'pending'"`
	// Status can be: pending, accepted, rejected, completed
}
