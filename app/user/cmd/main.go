package main

import (
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/config"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/handler"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/repository"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/routes"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/service"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/usecase"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	sharedConfig "github.com/PorcoGalliard/eCommerce-Microservice/pkg/config"
	"github.com/PorcoGalliard/eCommerce-Microservice/resource"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetupLogger()
	config := sharedConfig.LoadConfig(&config.Config{},
		sharedConfig.WithConfigPath("files/config"),
		sharedConfig.WithConfigFile("user_service_config"),
		sharedConfig.WithConfigType("yaml"),
	)
	
	postgres := resource.InitPostgres(config.Database)
	redis := resource.InitRedis(config.Redis)
	router := gin.Default()

	// Repository
	userRepository := repository.NewUserRepository(redis, postgres)

	// Service
	userService := service.NewUserService(userRepository)

	// Usecase
	userUsecase := usecase.NewUserUsecase(userService, config.Secret.JWTSecret)

	// Handler
	userHandler := handler.NewUserHandler(userUsecase)

	routes.SetupRoutes(router, userHandler, config.Secret.JWTSecret)
	router.Run(":" + config.App.Port)
}