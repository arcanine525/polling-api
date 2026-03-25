package database

import (
	"fmt"
	"log"

	"polling-system/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(
	// &authmodels.User{},
	// &pollmodels.Poll{},
	// &pollmodels.Candidate{},
	// &pollmodels.PollQuestion{},
	// &pollmodels.Answer{},
	// &contactmodels.File{},
	// &contactmodels.Contact{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	fmt.Println("Database connected (postgres)")
}
