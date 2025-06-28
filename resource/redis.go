package resource

import (
	"context"
	"fmt"

	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/pkg/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg config.RedisConfig) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Logger.Fatalf("❌ Failed connect to Redis: %v", err)
	}

	log.Logger.Printf("✅ Successfully connect to Redis with Host: %s and Port: %s", cfg.Host, cfg.Port)
	return redisClient
}