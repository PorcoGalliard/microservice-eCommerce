package usecase

import (
	// golang package
	"context"
	"errors"
	"fmt"
	"paymentfc/cmd/payment/service"
	"paymentfc/infrastructure/constant"
	"paymentfc/infrastructure/log"
	"paymentfc/models"
	"paymentfc/pdf"
	"strconv"
	"strings"
	"time"

	// external package
	"github.com/sirupsen/logrus"
)

type PaymentUsecase interface {
	// DownloadPDFInvoice download pdf invoice by given orderID.
	//
	// It returns string, and nil error when successful.
	// Otherwise, empty string, and error will be returned.
	DownloadPDFInvoice(ctx context.Context, orderID int64) (string, error)

	// FailedPaymentList failed payment list.
	//
	// It returns models.FailedPaymentList, and nil error when successful.
	// Otherwise, empty models.FailedPaymentList, and error will be returned.
	FailedPaymentList(ctx context.Context) (models.FailedPaymentList, error)

	// ProcessPaymentRequests process payment requests by given OrderCreatedEvent.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	ProcessPaymentRequests(ctx context.Context, payload models.OrderCreatedEvent) error

	// ProcessPaymentWebhook process payment webhook by given XenditWebhookPayload.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	ProcessPaymentWebhook(ctx context.Context, param models.XenditWebhookPayload) error
}

type paymentUsecase struct {
	Service service.PaymentService
}

// NewPaymentUsecase new payment usecase by given PaymentService.
//
// It returns PaymentUsecase when successful.
// Otherwise, empty PaymentUsecase will be returned.
func NewPaymentUsecase(svc service.PaymentService) PaymentUsecase {
	return &paymentUsecase{
		Service: svc,
	}
}

// DownloadPDFInvoice download pdf invoice by given orderID.
//
// It returns string, and nil error when successful.
// Otherwise, empty string, and error will be returned.
func (uc *paymentUsecase) DownloadPDFInvoice(ctx context.Context, orderID int64) (string, error) {
	paymentDetail, err := uc.Service.GetPaymentInfoByOrderID(ctx, orderID)
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("/fcproject/invoice_%d", orderID)
	err = pdf.GenerateInvoicePDF(paymentDetail, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// FailedPaymentList failed payment list.
//
// It returns models.FailedPaymentList, and nil error when successful.
// Otherwise, empty models.FailedPaymentList, and error will be returned.
func (uc *paymentUsecase) FailedPaymentList(ctx context.Context) (models.FailedPaymentList, error) {
	paymentList, err := uc.Service.GetFailedPaymentList(ctx)
	if err != nil {
		return models.FailedPaymentList{}, err
	}

	result := models.FailedPaymentList{
		TotalFailedPayment: len(paymentList),
		PaymentList:        paymentList,
	}

	return result, nil
}

// ProcessPaymentRequests process payment requests by given OrderCreatedEvent.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (uc *paymentUsecase) ProcessPaymentRequests(ctx context.Context, payload models.OrderCreatedEvent) error {
	err := uc.Service.SavePaymentRequests(ctx, models.PaymentRequests{
		OrderID:    payload.OrderID,
		Amount:     payload.TotalAmount,
		UserID:     payload.UserID,
		Status:     "PENDING",
		CreateTime: time.Now(),
	})

	if err != nil {
		return err
	}

	return nil
}

// ProcessPaymentWebhook process payment webhook by given XenditWebhookPayload.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (uc *paymentUsecase) ProcessPaymentWebhook(ctx context.Context, payload models.XenditWebhookPayload) error {
	switch payload.Status {
	case "PAID":
		// construct external id --> order id
		orderID := extractOrderID(payload.ExternalID)

		// validate webhook amount before process payment success
		amount, err := uc.Service.CheckPaymentAmountByOrderID(ctx, orderID)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"order_id":       orderID,
				"status":         payload.Status,
				"external_id":    payload.ExternalID,
				"webhook_amount": payload.Amount,
			})
			return err
		}

		if amount != payload.Amount {
			// insert to table payment anomalies
			errorMessage := fmt.Sprintf("Webhook amount mismatch: expected %.2f, got %.2f", amount, payload.Amount)
			paymentAnomaly := models.PaymentAnomaly{
				OrderID:     orderID,
				ExternalID:  payload.ExternalID,
				AnomalyType: constant.AnomalyTypeInvalidAmount,
				Notes:       errorMessage,
				Status:      constant.PaymentAnomalyStatusNeedToCheck,
				CreateTime:  time.Now(),
			}

			err := uc.Service.SavePaymentAnomaly(ctx, paymentAnomaly)
			if err != nil {
				log.Logger.WithFields(logrus.Fields{
					"payload":        payload,
					"paymentAnomaly": paymentAnomaly,
				}).WithError(err)
				return err
			}

			log.Logger.WithFields(logrus.Fields{
				"payload": payload,
			}).Errorf("Webhook amount mismatch: expected %.2f, got %.2f", amount, payload.Amount)
			err = errors.New(errorMessage)
			return err
		}

		// connect ke service layer kita
		err = uc.Service.ProcessPaymentSuccess(ctx, orderID)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"status":      payload.Status,
				"external_id": payload.ExternalID,
			}).Errorf("uc.Service.ProcessPaymentSuccess() got error: %v", err)
			return err
		}
	case "FAILED":
		orderID := extractOrderID(payload.ExternalID)
		err := uc.Service.ProcessPaymentFailed(ctx, orderID)
		if err != nil {
			return err
		}
	case "PENDING":
		//
	default:
		log.Logger.WithFields(logrus.Fields{
			"status":      payload.Status,
			"external_id": payload.ExternalID,
		}).Infof("[%s] Anomaly Payment Webhook Status: %s", payload.ExternalID, payload.Status)

		// next kita akan buat table baru 'payment_anomaly'
	}

	return nil
}

// extractOrderID extract order id by given externalID.
//
// It returns int64 when successful.
// Otherwise, empty int64 will be returned.
func extractOrderID(externalID string) int64 {
	// order id: 12345
	// key kafka event: "order-12345"
	idStr := strings.TrimPrefix(externalID, "order-")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	return id
}