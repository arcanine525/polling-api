package handlers

import (
	"net/http"

	"polling-system/auth/models"
	"polling-system/auth/services"
	"polling-system/database"

	"github.com/gin-gonic/gin"
)

// firebaseAPIKey is set from main.go during route setup.
var firebaseAPIKey string

// SetFirebaseAPIKey stores the API key for use by login handler.
func SetFirebaseAPIKey(key string) {
	firebaseAPIKey = key
}

// Login handles POST /api/v1/auth/login
// @Summary      Login with email and password
// @Description  Authenticates a user with email/password via Firebase Auth, provisions the user if new, and returns the user profile with tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  models.LoginRequest  true  "Login credentials"
// @Success      200  {object}  models.LoginResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate with Firebase Auth REST API
	fbResp, err := services.SignInWithEmailPassword(firebaseAPIKey, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Provision user in database
	user, err := services.ProvisionUser(database.DB, models.ProvisionRequest{
		FirebaseUID: fbResp.LocalID,
		Email:       fbResp.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision user"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		User:         *user,
		IDToken:      fbResp.IDToken,
		RefreshToken: fbResp.RefreshToken,
		ExpiresIn:    fbResp.ExpiresIn,
	})
}

// TokenExchange handles POST /api/v1/auth/token
// @Summary      Exchange Firebase token for user profile
// @Description  Verifies the Firebase ID token (via middleware), provisions the user if new, and returns the user profile.
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /auth/token [post]
func TokenExchange(c *gin.Context) {
	uid := c.GetString("firebase_uid")
	email := c.GetString("firebase_email")

	user, err := services.ProvisionUser(database.DB, models.ProvisionRequest{
		FirebaseUID: uid,
		Email:       email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision user"})
		return
	}
	c.JSON(http.StatusOK, user)
}
