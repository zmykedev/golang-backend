package models

import (
	"time"

	"gorm.io/gorm"
)

type Tourist struct {
	gorm.Model
	UserID        uint      `json:"user_id" gorm:"not null"`
	User          User      `json:"user" gorm:"foreignKey:UserID"`
	Nationality   string    `json:"nationality" gorm:"not null"`
	Language      string    `json:"language" gorm:"not null"`
	ArrivalDate   time.Time `json:"arrival_date" gorm:"not null"`
	DepartureDate time.Time `json:"departure_date" gorm:"not null"`
	Preferences   string    `json:"preferences"`
	SpecialNeeds  string    `json:"special_needs"`
	Status        string    `json:"status" gorm:"default:'pending'"` // pending, active, completed
}
