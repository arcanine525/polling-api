package middleware

import (
	"net/http"

	"polling-system/auth/models"
	"polling-system/auth/services"
	"polling-system/database"

	"github.com/gin-gonic/gin"
)

// UserProvisioning upserts a users row on every authenticated request
// and sets user_id, user_role, and user in the Gin context.
// Must run after AuthRequired.
func UserProvisioning() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("firebase_uid")
		if uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "firebase_uid not found in context"})
			return
		}

		email := c.GetString("firebase_email")

		user, err := services.ProvisionUser(database.DB, models.ProvisionRequest{
			FirebaseUID: uid,
			Email:       email,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to provision user"})
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("user", user)

		c.Next()
	}
}
