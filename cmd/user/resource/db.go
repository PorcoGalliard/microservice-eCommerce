package resource

import (
	"fmt"
	"log"

	"github.com/PorcoGalliard/eCommerce-Microservice/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,cfg.Database.Name, cfg.Database.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Error connect to Postgres => %v", err)
	}

	log.Println("Connected to Postgres")
	return db
} 