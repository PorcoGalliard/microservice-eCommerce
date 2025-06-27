package routes

import (
	"github.com/PorcoGalliard/eCommerce-Microservice/cmd/user/handler"
	"github.com/PorcoGalliard/eCommerce-Microservice/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userHandler handler.UserHandler) {
	router.Use(middleware.RequestLogger())
	router.GET("/ping", userHandler.Ping)
	router.POST("/v1/register", userHandler.Register)
	router.POST("/v1/login", userHandler.Login)
}