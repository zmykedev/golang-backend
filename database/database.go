package database

import (
	"log"
	"os"

	"fiber-backend/models"
	"fiber-backend/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create the enum type if it doesn't exist
	db.Exec(`DO $$ BEGIN
		CREATE TYPE user_role AS ENUM ('tourist', 'driver', 'admin');
		EXCEPTION WHEN duplicate_object THEN null;
	END $$;`)

	// Auto Migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Tourist{},
		&models.Booking{},
		&models.TouristRequest{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Ensure google_id is nullable
	if err := db.Exec(`ALTER TABLE users ALTER COLUMN google_id DROP NOT NULL;`).Error; err != nil {
		utils.LogError("Failed to make google_id nullable: %v", err)
	}

	DB = db
	utils.LogInfo("Database connected and migrated successfully")
}
