package handler

import (
	// golang package
	"net/http"
	"paymentfc/cmd/payment/usecase"
	"paymentfc/infrastructure/log"
	"paymentfc/models"
	"strconv"

	// external package
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PaymentHandler interface {
	// HandleDownloadPDFInvoice handle download pdf invoice by given c pointer of gin.Context.
	HandleDownloadPDFInvoice(c *gin.Context)

	// HandleXenditWebhook handle xendit webhook by given c pointer of gin.Context.
	HandleXenditWebhook(c *gin.Context)

	// HandleCreateInvoice handle create invoice by given c pointer of gin.Context.
	HandleCreateInvoice(c *gin.Context)

	// HandleFailedPayments handle failed payments by given c pointer of gin.Context.
	HandleFailedPayments(c *gin.Context)
}

type paymentHandler struct {
	Usecase            usecase.PaymentUsecase
	XenditWebhookToken string
}

// NewPaymentHandler new payment handler by given PaymentUsecase.
//
// It returns PaymentHandler when successful.
// Otherwise, empty PaymentHandler will be returned.
func NewPaymentHandler(usecase usecase.PaymentUsecase, webhookToken string) PaymentHandler {
	return &paymentHandler{
		Usecase:            usecase,
		XenditWebhookToken: webhookToken,
	}
}

// HandleFailedPayments handle failed payments by given c pointer of gin.Context.
func (h *paymentHandler) HandleFailedPayments(c *gin.Context) {
	failedPaymentList, err := h.Usecase.FailedPaymentList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	// push to dashboard monitoring
	// push ke channel
	c.JSON(http.StatusOK, gin.H{
		"data": failedPaymentList,
	})
}

// HandleCreateInvoice handle create invoice by given c pointer of gin.Context.
func (h *paymentHandler) HandleCreateInvoice(c *gin.Context) {
	var payload models.OrderCreatedEvent
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	return
}

// HandleXenditWebhook handle xendit webhook by given c pointer of gin.Context.
func (h *paymentHandler) HandleXenditWebhook(c *gin.Context) {
	var payload models.XenditWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"payload": payload,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "error_message": err.Error()})
		return
	}

	// validate webhook token
	headerWebhookToken := c.GetHeader("x-callback-token")
	if h.XenditWebhookToken != headerWebhookToken {
		log.Logger.WithFields(logrus.Fields{
			"callbackToken": headerWebhookToken,
		}).Errorf("Invalid Webhook Token: %s", headerWebhookToken)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid webhook token!"})
		return
	}

	err := h.Usecase.ProcessPaymentWebhook(c.Request.Context(), payload)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"payload": payload,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success!",
	})
	return
}

// HandleDownloadPDFInvoice handle download pdf invoice by given c pointer of gin.Context.
func (h *paymentHandler) HandleDownloadPDFInvoice(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)

	filePath, err := h.Usecase.DownloadPDFInvoice(c.Request.Context(), orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).WithError(err).Errorf("h.Usecase.DownloadPDFInvoice() got error: %v", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": err.Error(),
		})
		return
	}

	c.FileAttachment(filePath, filePath)
}