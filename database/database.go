package database

import (
	"fmt"
	"log"
	"os"

	"fiber-backend/models"
	"fiber-backend/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Drop the role column if it exists
	if err := db.Exec(`ALTER TABLE IF EXISTS users DROP COLUMN IF EXISTS role;`).Error; err != nil {
		utils.LogError("Failed to drop role column: %v", err)
	}

	// Create the enum type
	if err := db.Exec(`DROP TYPE IF EXISTS user_role;`).Error; err != nil {
		utils.LogError("Failed to drop existing enum type: %v", err)
	}

	if err := db.Exec(`CREATE TYPE user_role AS ENUM ('tourist', 'driver', 'admin');`).Error; err != nil {
		log.Fatal("Failed to create enum type:", err)
	}

	// Auto Migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Tourist{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Add the role column with the correct type
	if err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS role user_role DEFAULT NULL;`).Error; err != nil {
		log.Fatal("Failed to add role column:", err)
	}

	DB = db
	utils.LogInfo("Database connected and migrated successfully")
}
