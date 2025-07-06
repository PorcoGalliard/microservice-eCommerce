package routes

import (
	"github.com/PorcoGalliard/eCommerce-Microservice/app/product/handler"
	"github.com/PorcoGalliard/eCommerce-Microservice/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, productHandler *handler.ProductHandler) {
	router.Use(middleware.RequestLogger())
	router.POST("/v1/product_category", productHandler.ProductCategoryManagement)
	router.POST("/v1/product", productHandler.ProductManagement)

	router.GET("/v1/product/:id", productHandler.GetProductInfo)
	router.GET("/v1/product_category/:id", productHandler.GetProductCategoryInfo)
}