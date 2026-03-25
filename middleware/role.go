package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RoleRequired checks that the authenticated user's role is one of the allowed roles.
// Must run after UserProvisioning (which sets "user_role" in context).
func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")

		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("forbidden: requires one of [%s]", strings.Join(allowedRoles, ", ")),
		})
	}
}
