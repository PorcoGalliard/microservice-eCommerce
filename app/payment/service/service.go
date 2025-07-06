package service

import (
	// golang package
	"context"
	"math"
	"paymentfc/cmd/payment/repository"
	"paymentfc/infrastructure/constant"
	"paymentfc/infrastructure/log"
	"paymentfc/models"
	"time"

	// external package
	"github.com/sirupsen/logrus"
)

const (
	maxRetryPublish = 5
)

// mockgen -source=cmd/payment/service/payment_service.go -destination=cmd/test_mocks/payment_service_mock.go -package=mocks
type PaymentService interface {
	// CheckPaymentAmountByOrderID check payment amount by order id by given orderID.
	//
	// It returns float64, and nil error when successful.
	// Otherwise, empty float64, and error will be returned.
	CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error)

	// GetFailedPaymentList get failed payment list.
	//
	// It returns slice of models.PaymentRequests, and nil error when successful.
	// Otherwise, nil value of models.PaymentRequests slice, and error will be returned.
	GetFailedPaymentList(ctx context.Context) ([]models.PaymentRequests, error)

	// GetPaymentInfoByOrderID get payment info by order id by given orderID.
	//
	// It returns models.Payment, and nil error when successful.
	// Otherwise, empty models.Payment, and error will be returned.
	GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (models.Payment, error)

	// ProcessPaymentFailed process payment failed by given orderID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	ProcessPaymentFailed(ctx context.Context, orderID int64) error

	// ProcessPaymentSuccess process payment success by given orderID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	ProcessPaymentSuccess(ctx context.Context, orderID int64) error

	// SavePaymentAnomaly save payment anomaly by given PaymentAnomaly.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error

	// SavePaymentRequests save payment requests by given PaymentRequests.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	SavePaymentRequests(ctx context.Context, param models.PaymentRequests) error
}

type paymentService struct {
	database  repository.PaymentDatabase
	publisher repository.PaymentEventPublisher
}

// NewPaymentService new payment service by given PaymentDatabase, and PaymentEventPublisher.
//
// It returns PaymentService when successful.
// Otherwise, empty PaymentService will be returned.
func NewPaymentService(database repository.PaymentDatabase, publisher repository.PaymentEventPublisher) PaymentService {
	return &paymentService{
		database:  database,
		publisher: publisher,
	}
}

// GetFailedPaymentList get failed payment list.
//
// It returns slice of models.PaymentRequests, and nil error when successful.
// Otherwise, nil value of models.PaymentRequests slice, and error will be returned.
func (s paymentService) GetFailedPaymentList(ctx context.Context) ([]models.PaymentRequests, error) {
	paymentList, err := s.database.GetFailedPaymentList(ctx)
	if err != nil {
		return nil, err
	}

	return paymentList, nil
}

// CheckPaymentAmountByOrderID check payment amount by order id by given orderID.
//
// It returns float64, and nil error when successful.
// Otherwise, empty float64, and error will be returned.
func (s paymentService) CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error) {
	amount, err := s.database.CheckPaymentAmountByOrderID(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.CheckPaymentAmountByOrderID() got error: %v", err)
		return 0, err
	}

	return amount, nil
}

// GetPaymentInfoByOrderID get payment info by order id by given orderID.
//
// It returns models.Payment, and nil error when successful.
// Otherwise, empty models.Payment, and error will be returned.
func (s paymentService) GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (models.Payment, error) {
	paymentInfo, err := s.database.GetPaymentInfoByOrderID(ctx, orderID)
	if err != nil {
		return models.Payment{}, err
	}

	return paymentInfo, nil
}

// SavePaymentAnomaly save payment anomaly by given PaymentAnomaly.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (s paymentService) SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error {
	err := s.database.SavePaymentAnomaly(ctx, param)
	if err != nil {
		return err
	}

	return nil
}

// SavePaymentRequests save payment requests by given PaymentRequests.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (s paymentService) SavePaymentRequests(ctx context.Context, param models.PaymentRequests) error {
	err := s.database.SavePaymentRequests(ctx, param)
	if err != nil {
		return err
	}

	return nil
}

// ProcessPaymentFailed process payment failed by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (s paymentService) ProcessPaymentFailed(ctx context.Context, orderID int64) error {
	// check payment info apakah status udah failed?
	paymentInfo, err := s.database.GetPaymentInfoByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	if paymentInfo.Status == "FAILED" {
		return nil // skip process
	}

	// publish event payment status
	err = retryPublishPayment(maxRetryPublish, func() error {
		// insert audit log
		auditLogParam := models.PaymentAuditLog{
			OrderID:    orderID,
			Event:      "PublishEventPaymentStatus-Failed",
			Actor:      "payment",
			CreateTime: time.Now(),
		}
		errAuditLog := s.database.InsertAuditLog(ctx, auditLogParam)
		if errAuditLog != nil {
			log.Logger.WithFields(logrus.Fields{
				"param": auditLogParam,
			}).WithError(errAuditLog).Errorf("s.database.InsertAuditLog() got error: %v", errAuditLog)
		}

		// push notifier
		return s.publisher.PublishEventPaymentStatus(ctx, orderID, "FAILED", "payment.failed")
	})

	// update status db
	err = s.database.MarkFailed(ctx, orderID)
	if err != nil {
		return err
	}

	return nil
}

// ProcessPaymentSuccess process payment success by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (s paymentService) ProcessPaymentSuccess(ctx context.Context, orderID int64) error {
	// validate either order id already executed
	isAlreadyPaid, err := s.database.IsAlreadyPaid(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.IsAlreadyPaid() got error: %v", err)
		return err
	}

	if isAlreadyPaid {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Infof("[skip - order %d] Payment status already paid!", orderID)
		return nil
	}

	// implement retry mechanism

	// publish event kafka
	// retry until max retry publish
	err = retryPublishPayment(maxRetryPublish, func() error {
		errLogAudit := s.database.InsertAuditLog(ctx, models.PaymentAuditLog{
			OrderID:    orderID,
			Event:      "PublishEventPaymentStatus",
			Actor:      "payment",
			CreateTime: time.Now(),
		})
		if errLogAudit != nil {
			log.Logger.WithFields(logrus.Fields{
				"order_id": orderID,
				"event":    "PublishEventPaymentStatus",
				"actor":    "payment",
			}).WithError(errLogAudit).Errorf("s.database.InsertAuditLog() got error: %v", errLogAudit)
		}

		return s.publisher.PublishEventPaymentStatus(ctx, orderID, "PAID", "payment.success")
	})
	if err != nil {
		// store data to DB --> failed_events
		failedEventsParam := models.FailedEvents{
			OrderID:    orderID,
			FailedType: constant.FailedPublishEventPaymentSuccess,
			Status:     constant.FailedPublishEventStatusNeedToCheck,
			Notes:      err.Error(),
			CreateTime: time.Now(),
		}

		// dead letter table
		errSaveFailedPublish := s.database.SaveFailedPublishEvent(ctx, failedEventsParam)
		if errSaveFailedPublish != nil {
			log.Logger.WithFields(logrus.Fields{
				"failedEventsParam": failedEventsParam,
			}).WithError(errSaveFailedPublish)
			return errSaveFailedPublish
		}

		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.publisher.PublishPaymentSuccess() got error: %v", err)
		return err
	}

	// update status DB
	err = s.database.MarkPaid(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.MarkPaid() got error: %v", err)
		return err
	} else {
		errLogAudit := s.database.InsertAuditLog(ctx, models.PaymentAuditLog{
			OrderID:    orderID,
			Event:      "MarkPaid",
			Actor:      "payment",
			CreateTime: time.Now(),
		})
		if errLogAudit != nil {
			log.Logger.WithFields(logrus.Fields{
				"order_id": orderID,
				"event":    "MarkPaid",
				"actor":    "payment",
			}).WithError(errLogAudit).Errorf("s.database.InsertAuditLog() got error: %v", errLogAudit)
		}
	}

	return nil
}

// retryPublishPayment retry publish payment by given max, and fn.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func retryPublishPayment(max int, fn func() error) error {
	var err error
	for i := range max {
		err = fn()
		if err == nil {
			return nil
		}

		// publish event
		// failed --> retry
		// set jeda (2 seconds)
		// failed --> retry
		wait := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Logger.Printf("Retry: %d, Error: %s. Retrying in %d seconds...", i+1, err, wait)
		time.Sleep(wait)
	}

	return err
}

package service

import (
	// golang package
	"context"
	"fmt"
	"paymentfc/cmd/payment/repository"
	"paymentfc/grpc"
	"paymentfc/infrastructure/log"
	"paymentfc/models"
	"time"

	// external package
	"github.com/sirupsen/logrus"
)

type SchedulerService struct {
	UserClient     grpc.UserClient
	Database       repository.PaymentDatabase
	Xendit         repository.XenditClient
	Publisher      repository.PaymentEventPublisher
	PaymentService PaymentService
}

// StartSweepingExpiredPendingPayments start sweeping expired pending payments.
func (s *SchedulerService) StartSweepingExpiredPendingPayments() {
	go func(ctx context.Context) {
		for {
			log.Logger.Println("Scheduler StartSweepingExpiredPendingPayments is running...")
			expiredPayments, err := s.Database.GetExpiredPendingPayments(ctx)
			if err != nil {
				log.Logger.Println("Failed get expired pending payments, err: ", err.Error())
				time.Sleep(5 * time.Minute) // optional
				continue
			}

			for _, expiredPayment := range expiredPayments {
				// check payment status first before update
				paymentInfo, err := s.Database.GetPaymentInfoByOrderID(ctx, expiredPayment.OrderID)
				if err != nil {
					log.Logger.Printf("[payment id: %d] Failed Get Payment Info, err: %s", expiredPayment.ID, err.Error())
				}

				// Status: Expired, Success, Failed --> ignore
				if paymentInfo.Status != "PENDING" {
					continue
				}

				// publish event payment failed
				err = retryPublishPayment(maxRetryPublish, func() error {
					return s.Publisher.PublishEventPaymentStatus(ctx, expiredPayment.OrderID, "FAILED", "payment.failed")
				})

				// mark expired
				err = s.Database.MarkExpired(ctx, expiredPayment.ID)
				if err != nil {
					log.Logger.Printf("[payment id: %d] Failed update expired, err: %s", expiredPayment.ID, err.Error())
				}
			}

			time.Sleep(10 * time.Minute)
		}
	}(context.Background())
}

// StartProcessFailedPaymentRequests start process failed payment requests.
func (s *SchedulerService) StartProcessFailedPaymentRequests() {
	go func(ctx context.Context) {
		for {
			// get list of failed payment requests from DB
			var paymentRequests []models.PaymentRequests
			err := s.Database.GetFailedPaymentRequests(ctx, &paymentRequests)
			if err != nil {
				log.Logger.Println("Error get failed payment requests! error: ", err.Error())

				time.Sleep(10 * time.Second)
				continue
			}

			for _, paymentRequest := range paymentRequests {
				// update status menjadi pending
				err = s.Database.UpdatePendingPaymentRequests(ctx, paymentRequest.ID)
				if err != nil {
					log.Logger.Println("s.Database.UpdatePendingPaymentRequests() got error: ", err.Error())

					// menambah retry count
					errUpdateStatus := s.Database.UpdateFailedPaymentRequests(ctx, paymentRequest.ID, err.Error())
					if errUpdateStatus != nil {
						log.Logger.Println("s.Database.UpdateFailedPaymentRequests() got error: ", errUpdateStatus.Error())
					}

					continue
				}
			}

			time.Sleep(1 * time.Minute)
		}

	}(context.Background())
}

// StartProcessPendingPaymentRequests start process pending payment requests.
func (s *SchedulerService) StartProcessPendingPaymentRequests() {
	go func(ctx context.Context) {
		for {

			// get pending payment requests
			var paymentRequests []models.PaymentRequests
			err := s.Database.GetPendingPaymentRequests(ctx, &paymentRequests)
			if err != nil {
				log.Logger.Println("s.Database.GetPendingPaymentRequests() got error: ", err.Error())
				// kasih jeda (considering ada issue di DB)
				time.Sleep(10 * time.Second)
				continue
			}

			// looping list of pending payment requests
			for _, paymentRequest := range paymentRequests {
				log.Logger.Printf("[DEBUG] Processing Payment Request Order %d", paymentRequest.OrderID)

				// pengecekan apakah create invoice sudah pernah direquest.
				paymentInfo, err := s.Database.GetPaymentInfoByOrderID(ctx, paymentRequest.OrderID)
				if err != nil {
					log.Logger.Println("s.Database.GetPaymentInfoByOrderID() got error ", err.Error())
					continue
				}

				externalID := fmt.Sprintf("order-%d", paymentRequest.OrderID)
				if paymentInfo.ID != 0 {
					// update status payment request success
					err = s.Database.UpdateSuccessPaymentRequests(ctx, paymentRequest.ID)
					if err != nil {
						// to do: need to handle.
						log.Logger.Printf("[req id: %d] s.Database.UpdateSuccessPaymentRequets() got error: %s", paymentRequest.ID, err.Error())
					}

					continue
				}

				// get user info by grpc
				userInfo, err := s.UserClient.GetUserInfoByUserID(ctx, paymentInfo.UserID)
				if err != nil {
					log.Logger.WithFields(logrus.Fields{
						"user_id":    paymentInfo.UserID,
						"payment_id": paymentInfo.ID,
					}).WithError(err).Errorf("[req id: %d] s.UserClient.GetUserInfoByUserID() got error: %v", paymentInfo.ID, err)
					continue
				}

				userEmail := userInfo.Email
				xenditInvoiceRequestParam := models.XenditInvoiceRequest{
					ExternalID:  externalID,
					Amount:      paymentRequest.Amount,
					Description: fmt.Sprintf("[FC] Pembayaran Order %d", paymentRequest.OrderID),
					PayerEmail:  userEmail,
				}

				xenditInvoiceDetail, err := s.Xendit.CreateInvoice(ctx, xenditInvoiceRequestParam)

				// get audit log by order id (order by create time)
				// audit log akan munculin semua action / event di order id tersebut
				paymentAuditLogParam := models.PaymentAuditLog{
					OrderID:    paymentInfo.OrderID,
					UserID:     paymentInfo.UserID,
					PaymentID:  paymentInfo.ID,
					ExternalID: externalID,
					Event:      "CreateInvoice",
					Actor:      "xendit",
					CreateTime: time.Now(),
				}

				errAuditLog := s.Database.InsertAuditLog(ctx, paymentAuditLogParam)
				if errAuditLog != nil {
					log.Logger.WithFields(logrus.Fields{
						"auditLogParam": paymentAuditLogParam,
					}).WithError(errAuditLog).Errorf("s.Database.InsertAuditLog() got error %v", errAuditLog)
				}

				if err != nil {
					log.Logger.Printf("[req id: %d] s.Xendit.CreateInvoice() got error: %v", paymentRequest.ID, err.Error())

					errSaveFailedPaymentRequest := s.Database.UpdateFailedPaymentRequests(ctx, paymentRequest.ID, err.Error())
					if errSaveFailedPaymentRequest != nil {
						log.Logger.Printf("[req id: %d] s.Database.UpdateFailedPaymentRequests() got error: %v", paymentRequest.ID, errSaveFailedPaymentRequest.Error())
					}

					continue
				}

				// update status payment request success
				err = s.Database.UpdateSuccessPaymentRequests(ctx, paymentRequest.ID)
				if err != nil {
					// to do: need to handle.
					log.Logger.Printf("[req id: %d] s.Database.UpdateSuccessPaymentRequets() got error: %s", paymentRequest.ID, err.Error())
				}

				// save data to table 'payments'
				err = s.Database.SavePayment(ctx, models.Payment{
					OrderID:     paymentRequest.OrderID,
					UserID:      paymentRequest.UserID,
					Amount:      paymentRequest.Amount,
					ExternalID:  externalID,
					Status:      "PENDING",
					CreateTime:  time.Now(),
					ExpiredTime: xenditInvoiceDetail.ExpiryDate,
				})
				if err != nil {
					// to do: need to handle.
					log.Logger.Printf("[req id: %d] s.Database.SavePayment() got error: %s", paymentRequest.ID, err.Error())
				}
			}

			time.Sleep(5 * time.Second) // jeda 5 detik per setiap polling
		}
	}(context.Background())
}

// StartCheckPendingInvoices start check pending invoices.
func (s *SchedulerService) StartCheckPendingInvoices() {
	ticker := time.NewTicker(10 * time.Minute)

	go func() {
		for range ticker.C {
			// query ke DB --> get list of pending invoices
			ctx := context.Background()
			listPendingInvoices, err := s.Database.GetPendingInvoices(ctx)
			if err != nil {
				log.Logger.Println("s.Database.GetPendingInvoices() got error: ", err.Error())
				continue
			}

			// looping dari hasil query
			for _, pendingInvoice := range listPendingInvoices {
				// iterate 1 per 1 dan execute utk cek status dengan hit ke endpoint xendit
				invoiceStatus, err := s.Xendit.CheckInvoiceStatus(ctx, pendingInvoice.ExternalID)
				if err != nil {
					log.Logger.Println("s.Xendit.CheckInvoiceStatus() got error: ", err.Error())
					continue
				}

				if invoiceStatus == "PAID" {
					err = s.PaymentService.ProcessPaymentSuccess(ctx, pendingInvoice.OrderID)
					if err != nil {
						log.Logger.Println("s.PaymentService.ProcessPaymentSuccess() got error: ", err)
						continue
					}
				}
				//..
			}
		}
	}()
}

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