package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProductRepository struct {
	Database *gorm.DB
	Redis *redis.Client
}

func NewProductRepository(db *gorm.DB, redis *redis.Client) *ProductRepository {
	return &ProductRepository{
		Database: db,
		Redis: redis,
	}
}