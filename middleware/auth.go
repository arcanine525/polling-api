package middleware

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var authClient *auth.Client

// InitFirebase initializes the Firebase Admin SDK and stores the auth client.
func InitFirebase(credentialsPath string) error {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}
	authClient, err = app.Auth(context.Background())
	return err
}

// AuthRequired verifies the Firebase ID token from the Authorization header.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Store Firebase claims in context
		c.Set("firebase_uid", token.UID)

		if email, ok := token.Claims["email"].(string); ok {
			c.Set("firebase_email", email)
		} else {
			c.Set("firebase_email", "")
		}

		if role, ok := token.Claims["role"].(string); ok {
			c.Set("firebase_role", role)
		} else {
			c.Set("firebase_role", "")
		}

		c.Next()
	}
}
