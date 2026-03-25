package contact

import (
	"polling-system/config"
	"polling-system/contact/handlers"
	"polling-system/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all contact/file routes on the given router group.
func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	// Files
	rg.POST("/files/upload", middleware.RoleRequired("admin", "poller"), handlers.UploadCSV(cfg))
	rg.GET("/files", handlers.ListFiles)

	// Contacts
	contacts := rg.Group("/contacts")
	{
		contacts.POST("", middleware.RoleRequired("admin", "poller"), handlers.CreateContact)
		contacts.POST("/bulk", middleware.RoleRequired("admin", "poller"), handlers.BulkCreateContacts)
		contacts.GET("", handlers.ListContacts)
		contacts.PATCH("/:id", middleware.RoleRequired("admin", "poller"), handlers.UpdateContact)
		contacts.DELETE("/:id", middleware.RoleRequired("admin"), handlers.DeleteContact)
	}
}
