package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseDSN             string
	Port                    string
	FirebaseCredentialsPath string
	FirebaseAPIKey          string
	UploadDir               string
}

func Load() *Config {
	cfg := &Config{
		DatabaseDSN:             getEnv("DB_DSN", "postgresql://localhost:5432/polling?sslmode=disable"),
		Port:                    getEnv("PORT", "8080"),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./serviceAccountKey.json"),
		FirebaseAPIKey:          getEnv("FIREBASE_API_KEY", ""),
		UploadDir:               getEnv("UPLOAD_DIR", "./uploads"),
	}

	if cfg.FirebaseAPIKey == "" {
		log.Println("WARNING: FIREBASE_API_KEY is not set — email/password login will not work")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
