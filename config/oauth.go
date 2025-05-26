package config

import (
	"fiber-backend/utils"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig *oauth2.Config

func InitOAuth() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")

	// Debug logging
	utils.LogInfo("Initializing OAuth with:")
	utils.LogInfo("Client ID: %s", clientID)
	utils.LogInfo("Callback URL: %s", callbackURL)
	if clientSecret == "" {
		utils.LogError("Google Client Secret is missing!")
	} else {
		utils.LogInfo("Client Secret is configured")
	}

	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
