package main

import (
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/config"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/handler"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/repository"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/routes"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/service"
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/usecase"
	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	sharedConfig "github.com/PorcoGalliard/eCommerce-Microservice/pkg/config"
	"github.com/PorcoGalliard/eCommerce-Microservice/resource"
	"github.com/gin-gonic/gin"
)

func main()  {
	log.SetupLogger()
	cfg := sharedConfig.LoadConfig(&config.ProductConfig{}, 
			sharedConfig.WithConfigPath("files/config"),
			sharedConfig.WithConfigFile("product_service_config"),
			sharedConfig.WithConfigType("yaml"))

	postgre := resource.InitPostgres(cfg.Database) 
	redis := resource.InitRedis(cfg.Redis)

	productRepository := repository.NewProductRepository(postgre, redis)
	productService := service.NewProductService(productRepository)
	productUsecase := usecase.NewProductUsecase(productService)
	productHandler := handler.NewProductHandler(productUsecase)

	router := gin.Default()
	routes.SetupRoutes(router, productHandler)

	router.Run(":"+cfg.App.Port)
}