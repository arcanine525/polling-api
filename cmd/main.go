package main

import (
	"fmt"
	"log"

	"polling-system/auth"
	authhandlers "polling-system/auth/handlers"
	"polling-system/config"
	"polling-system/contact"
	"polling-system/database"
	_ "polling-system/docs"
	"polling-system/middleware"
	"polling-system/polling"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title          Polling System API
// @version        1.0
// @description    Unified API for authentication, polls, contacts, and file management.

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your token in the format: Bearer {token}

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	database.Init(cfg)

	// Initialize Firebase Auth
	if err := middleware.InitFirebase(cfg.FirebaseCredentialsPath); err != nil {
		log.Fatalf("failed to init firebase: %v", err)
	}

	// Pass Firebase API key to auth handlers
	authhandlers.SetFirebaseAPIKey(cfg.FirebaseAPIKey)

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes (no auth)
	v1Public := r.Group("/api/v1")
	auth.RegisterPublicRoutes(v1Public)

	// Authenticated routes
	v1 := r.Group("/api/v1")
	v1.Use(middleware.AuthRequired())
	v1.Use(middleware.UserProvisioning())

	auth.RegisterRoutes(v1)
	polling.RegisterRoutes(v1)
	contact.RegisterRoutes(v1, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
