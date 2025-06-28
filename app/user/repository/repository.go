package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepository struct {
	Redis *redis.Client
	Database *gorm.DB
}

func NewUserRepository(redis *redis.Client, db *gorm.DB) *UserRepository {
	return &UserRepository{
		Redis: redis,
		Database: db,
	}
}