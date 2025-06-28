package routes

import (
	"github.com/PorcoGalliard/eCommerce-Microservice/app/user/handler"
	"github.com/PorcoGalliard/eCommerce-Microservice/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userHandler *handler.UserHandler, JWTSecret string) {
	// Public API
	router.Use(middleware.RequestLogger())
	router.POST("/v1/register", userHandler.Register)
	router.POST("/v1/login", userHandler.Login)
	router.GET("/v1/ping", userHandler.Ping)

	// Private API
	private := router.Group("/auth")
	private.Use(middleware.AuthMiddleware(JWTSecret))
	private.GET("/v1/user_info", userHandler.GetUserInfo)

}