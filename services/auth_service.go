package services

import (
	"encoding/json"
	"fiber-backend/models"
	"fiber-backend/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func (s *AuthService) HandleGoogleAuth(code string) (*models.User, string, error) {
	// Exchange code for tokens
	token, err := s.getGoogleToken(code)
	if err != nil {
		return nil, "", err
	}

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(token)
	if err != nil {
		return nil, "", err
	}

	// Find or create user
	user, err := s.findOrCreateGoogleUser(userInfo)
	if err != nil {
		return nil, "", err
	}

	// Generate JWT
	jwtToken, err := utils.GenerateJWT(user)
	if err != nil {
		return nil, "", err
	}

	return user, jwtToken, nil
}

func (s *AuthService) getGoogleToken(code string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	url := "https://oauth2.googleapis.com/token"
	data := fmt.Sprintf(
		"code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code",
		code, clientID, clientSecret, redirectURI,
	)

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["access_token"].(string), nil
}

func (s *AuthService) getGoogleUserInfo(token string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *AuthService) findOrCreateGoogleUser(userInfo *GoogleUserInfo) (*models.User, error) {
	var user models.User

	// Try to find existing user by Google ID first
	result := s.db.Where("google_id = ?", userInfo.ID).First(&user)
	if result.Error == nil {
		return &user, nil
	}

	// If not found by Google ID, try to find by email
	result = s.db.Where("email = ?", userInfo.Email).First(&user)
	if result.Error == nil {
		// Update existing user with Google ID
		user.GoogleID = userInfo.ID
		if err := s.db.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	// Create new user if not found
	user = models.User{
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		GoogleID: userInfo.ID,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
