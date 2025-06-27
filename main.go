package main

import (
	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/handler"
	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/repository"
	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/resource"
	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/service"
	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/usecase"
	"github.com/PorcoGalliard/eCommerce-Microservice/config"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/PorcoGalliard/eCommerce-Microservice/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	// Setup
	log.SetupLogger()
	cfg := config.LoadConfig()
	redis := resource.InitRedis(&cfg)
	port := cfg.App.Port
	db := resource.InitDB(&cfg)
	
	// Repository
	userRepository := repository.NewUserRepository(redis, db)

	// Service
	userService := service.NewUserService(*userRepository)

	// Usecase
	userUsecase := usecase.NewUserUsecase(*userService, cfg.Secret.JWTSecret)

	// Handler
	userHandler := handler.NewUserHandler(*userUsecase)

	// Routes
	router := gin.Default()
	routes.SetupRoutes(router, *userHandler)

	// Running
	router.Run(":" + port)
	log.Logger.Printf("Server running on port: %s", port)
}