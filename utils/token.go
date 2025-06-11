package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fiber-backend/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GetJWTSecret returns the JWT secret from environment or panics if not set
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	return []byte(secret)
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	secret := GetJWTSecret()
	return token.SignedString(secret)
}

// GenerateRandomPassword generates a random password for Google OAuth users
func GenerateRandomPassword() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
