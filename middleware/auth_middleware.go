package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error_message": "Missing required auth",
			})
			ctx.Abort()
			return
		}

		tokenString := strings.Split(authHeader, " ")
		if len(tokenString) != 2 {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error_message": "Invalid token",
			})
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(tokenString[1], func (token *jwt.Token) (interface{}, error)  {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error_message": "Invalid token",
			})

			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error_message": "Invalid token",
			})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims["user_id"].(float64))
		ctx.Next()
	}
}