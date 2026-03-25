package auth

import (
	"polling-system/auth/handlers"
	"polling-system/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers routes that don't require authentication.
func RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/login", handlers.Login)
}

// RegisterRoutes registers authenticated auth/user routes.
func RegisterRoutes(rg *gin.RouterGroup) {
	// Token exchange (requires Bearer)
	rg.POST("/auth/token", handlers.TokenExchange)

	// Users (self)
	rg.GET("/users/me", handlers.GetMe)
	rg.PATCH("/users/me", handlers.UpdateMe)

	// Users (admin)
	adminUsers := rg.Group("/users")
	adminUsers.Use(middleware.RoleRequired("admin"))
	{
		adminUsers.GET("", handlers.ListUsers)
		adminUsers.GET("/:id", handlers.GetUser)
		adminUsers.POST("", handlers.CreateUser)
		adminUsers.PATCH("/:id", handlers.UpdateUser)
		adminUsers.PATCH("/:id/role", handlers.UpdateUserRole)
		adminUsers.DELETE("/:id", handlers.DeleteUser)
	}
}
