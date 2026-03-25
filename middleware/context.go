package middleware

import (
	"polling-system/auth/models"

	"github.com/gin-gonic/gin"
)

// GetUserID returns the authenticated user's database ID from the Gin context.
func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

// GetUserRole returns the authenticated user's role from the Gin context.
func GetUserRole(c *gin.Context) string {
	return c.GetString("user_role")
}

// GetFirebaseUID returns the authenticated user's Firebase UID from the Gin context.
func GetFirebaseUID(c *gin.Context) string {
	return c.GetString("firebase_uid")
}

// GetUser returns the full User struct from the Gin context.
func GetUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
