package middleware

import (
	"context"
	"time"

	"github.com/PorcoGalliard/eCommerce-Microservice/infrastructure/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func RequestLogger() gin.HandlerFunc {
	return func (c *gin.Context)  {
		requestID := uuid.New().String()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
		defer cancel()

		ctx := context.WithValue(timeoutCtx, "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		requestLog := logrus.Fields{
			"request_id": requestID,
			"method": c.Request.Method,
			"path": c.Request.URL.Path,
			"status": c.Writer.Status(),
			"latency": latency,
		}

		if c.Writer.Status() == 201 || c.Writer.Status() == 200 {
			log.Logger.WithFields(requestLog).Info("Request Success")
		} else {
			log.Logger.WithFields(requestLog).Info("Request Error")
		}
	}
}