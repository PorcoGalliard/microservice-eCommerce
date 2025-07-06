package handler

import (
	// golang package
	"net/http"
	"orderfc/cmd/order/usecase"
	"orderfc/infrastructure/log"
	"orderfc/models"
	"strconv"

	// external package
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	OrderUsecase usecase.OrderUsecase
}

// NewOrderHandler new orderhandler by given OrderUsecase.
//
// It returns pointer of OrderHandler when successful.
// Otherwise, nil pointer of OrderHandler will be returned.
func NewOrderHandler(orderUsecase usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{
		OrderUsecase: orderUsecase,
	}
}

// Checkout checkout by given c pointer of gin.Context.
func (h *OrderHandler) Checkout(c *gin.Context) {
	var param models.CheckoutRequest
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	if len(param.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items must not be empty"})
		return
	}

	userIDStr, isExist := c.Get("user_id")
	if !isExist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Unauthorized",
		})
		return
	}

	userID, ok := userIDStr.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Invalid user id",
		})
		return
	}

	param.UserID = int64(userID)
	if param.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session, please try to login again!"})
		return
	}

	orderID, err := h.OrderUsecase.CheckoutOrder(c.Request.Context(), &param)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("h.OrderUsecase.CheckoutOrder() got error %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"status":   "created",
	})
}

// GetOrderHistory get order history by given c pointer of gin.Context.
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	userIDStr, isExist := c.Get("user_id")
	if !isExist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Unauthorized",
		})
		return
	}

	userID, ok := userIDStr.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Invalid user id",
		})
		return
	}

	status, _ := strconv.Atoi(c.DefaultQuery("status", "0")) // 0 = all
	orderHistoryParam := models.OrderHistoryParam{
		UserID: int64(userID),
		Status: status,
	}

	histories, err := h.OrderUsecase.GetOrderHistory(c.Request.Context(), orderHistoryParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": histories,
	})
}