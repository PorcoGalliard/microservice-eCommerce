package repository

import (
	// golang package
	"context"
	"paymentfc/infrastructure/log"
	"paymentfc/models"
	"time"

	// external package
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PaymentDatabase interface {
	// CheckPaymentAmountByOrderID check payment amount by order id by given orderID.
	//
	// It returns float64, and nil error when successful.
	// Otherwise, empty float64, and error will be returned.
	CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error)

	// GetExpiredPendingPayments get expired pending payments.
	//
	// It returns slice of models.Payment, and nil error when successful.
	// Otherwise, nil value of models.Payment slice, and error will be returned.
	GetExpiredPendingPayments(ctx context.Context) ([]models.Payment, error)

	// GetFailedPaymentList get failed payment list.
	//
	// It returns slice of models.PaymentRequests, and nil error when successful.
	// Otherwise, nil value of models.PaymentRequests slice, and error will be returned.
	GetFailedPaymentList(ctx context.Context) ([]models.PaymentRequests, error)

	// GetFailedPaymentRequests get failed payment requests by given .
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	GetFailedPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error

	// GetPaymentInfoByOrderID get payment info by order id by given orderID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (models.Payment, error)

	// GetPendingInvoices get pending invoices.
	//
	// It returns slice of models.Payment, and nil error when successful.
	// Otherwise, nil value of models.Payment slice, and error will be returned.
	GetPendingInvoices(ctx context.Context) ([]models.Payment, error)

	// GetPendingPaymentRequests get pending payment requests by given .
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	GetPendingPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error

	// InsertAuditLog insert audit log by given PaymentAuditLog.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	InsertAuditLog(ctx context.Context, param models.PaymentAuditLog) error

	// IsAlreadyPaid is already paid by given orderID.
	//
	// It returns bool, and nil error when successful.
	// Otherwise, empty bool, and error will be returned.
	IsAlreadyPaid(ctx context.Context, orderID int64) (bool, error)

	// MarkExpired mark expired by given paymentID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	MarkExpired(ctx context.Context, paymentID int64) error

	// MarkFailed mark failed by given orderID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	MarkFailed(ctx context.Context, orderID int64) error

	// MarkPaid mark paid by given orderID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	MarkPaid(ctx context.Context, orderID int64) error

	// SaveFailedPublishEvent save failed publish event by given FailedEvents.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	SaveFailedPublishEvent(ctx context.Context, param models.FailedEvents) error

	// SavePayment save payment by given Payment.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	SavePayment(ctx context.Context, param models.Payment) error

	// SavePaymentAnomaly save payment anomaly by given PaymentAnomaly.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error

	// SavePaymentRequests save payment request by given PaymentRequests.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	SavePaymentRequests(ctx context.Context, param models.PaymentRequests) error

	// UpdateFailedPaymentRequests update failed payment requests by given paymentRequestID, and notes.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	UpdateFailedPaymentRequests(ctx context.Context, paymentRequestID int64, notes string) error

	// UpdatePendingPaymentRequests update pending payment requests by given paymentRequestID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	UpdatePendingPaymentRequests(ctx context.Context, paymentRequestID int64) error

	// UpdateSuccessPaymentRequests update success payment requests by given paymentRequestID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	UpdateSuccessPaymentRequests(ctx context.Context, paymentRequestID int64) error
}

type paymentDatabase struct {
	DB *gorm.DB
}

// NewPaymentDatabase new payment database by given db pointer of gorm.DB.
//
// It returns PaymentDatabase when successful.
// Otherwise, empty PaymentDatabase will be returned.
func NewPaymentDatabase(db *gorm.DB) PaymentDatabase {
	return &paymentDatabase{
		DB: db,
	}
}

// MarkPaid mark paid by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) MarkPaid(ctx context.Context, orderID int64) error {
	// update status DB menjadi "paid"
	err := r.DB.Model(&models.Payment{}).Table("payments").WithContext(ctx).Where("order_id = ?", orderID).Update("status", "paid").Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("r.DB.Update() got error: %v", err)
		return err
	}

	return nil
}

// MarkFailed mark failed by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) MarkFailed(ctx context.Context, orderID int64) error {
	err := r.DB.Model(&models.Payment{}).Table("payments").WithContext(ctx).Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":      "FAILED",
			"update_time": time.Now(),
		}).Error

	if err != nil {
		return err
	}

	return nil
}

// GetPaymentInfoByOrderID get payment info by order id by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (models.Payment, error) {
	var result models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("order_id = ?", orderID).First(&result).Error
	if err != nil {
		return models.Payment{}, err
	}

	return result, nil
}

// CheckPaymentAmountByOrderID check payment amount by order id by given orderID.
//
// It returns float64, and nil error when successful.
// Otherwise, empty float64, and error will be returned.
func (r *paymentDatabase) CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error) {
	var result models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("order_id = ?", orderID).First(&result).Error
	if err != nil {
		return 0, err
	}

	return result.Amount, nil
}

// GetPendingInvoices get pending invoices.
//
// It returns slice of models.Payment, and nil error when successful.
// Otherwise, nil value of models.Payment slice, and error will be returned.
func (r *paymentDatabase) GetPendingInvoices(ctx context.Context) ([]models.Payment, error) {
	var result []models.Payment
	// data di DB ada > 10mil data
	err := r.DB.Table("payments").WithContext(ctx).Where("status = ? AND create_time >= now() - interval '1 day'", "PENDING").Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

// IsAlreadyPaid is already paid by given orderID.
//
// It returns bool, and nil error when successful.
// Otherwise, empty bool, and error will be returned.
func (r *paymentDatabase) IsAlreadyPaid(ctx context.Context, orderID int64) (bool, error) {
	var result models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("order_id = ?", orderID).First(&result).Error
	if err != nil {
		return false, err
	}

	return result.Status == "PAID", nil
}

// SavePayment save payment by given Payment.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) SavePayment(ctx context.Context, param models.Payment) error {
	err := r.DB.Table("payments").WithContext(ctx).Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("r.DB.Create() got error: %v", err)
		return err
	}

	return nil
}

// SavePaymentAnomaly save payment anomaly by given PaymentAnomaly.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error {
	err := r.DB.Table("payment_anomalies").WithContext(ctx).Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("r.DB.Create() got error: %v", err)
		return err
	}

	return nil
}

// SaveFailedPublishEvent save failed publish event by given FailedEvents.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) SaveFailedPublishEvent(ctx context.Context, param models.FailedEvents) error {
	err := r.DB.Table("failed_events").WithContext(ctx).Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).WithError(err)
		return err
	}

	return nil
}

// SavePaymentRequests save payment requests by given PaymentRequests.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) SavePaymentRequests(ctx context.Context, param models.PaymentRequests) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Create(models.PaymentRequests{
		OrderID:    param.OrderID,
		UserID:     param.UserID,
		Amount:     param.Amount,
		Status:     param.Status,
		CreateTime: param.CreateTime,
	}).Error
	if err != nil {
		return err
	}

	return nil
}

// GetPendingPaymentRequests get pending payment requests by given .
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) GetPendingPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("status = ?", "PENDING").Limit(5).Order("create_time ASC").Find(paymentRequests).Error
	if err != nil {
		return err
	}

	return nil
}

// GetFailedPaymentRequests get failed payment requests by given .
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) GetFailedPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("status = ?", "FAILED").
		Where("retry_count <= ?", 3).Limit(5).Order("create_time ASC").Find(paymentRequests).Error
	if err != nil {
		return err
	}

	return nil
}

// GetFailedPaymentList get failed payment list.
//
// It returns slice of models.PaymentRequests, and nil error when successful.
// Otherwise, nil value of models.PaymentRequests slice, and error will be returned.
func (r *paymentDatabase) GetFailedPaymentList(ctx context.Context) ([]models.PaymentRequests, error) {
	var paymentList []models.PaymentRequests
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("status = ? AND retry_count >= ?", "FAILED", 3).Order("create_time ASC").Find(&paymentList).Error
	if err != nil {
		return nil, err
	}

	return paymentList, nil
}

// UpdateSuccessPaymentRequests update success payment requests by given paymentRequestID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) UpdateSuccessPaymentRequests(ctx context.Context, paymentRequestID int64) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("id = ?", paymentRequestID).
		Updates(map[string]interface{}{
			"status":      "SUCCESS",
			"update_time": time.Now(),
		}).Error
	if err != nil {
		return err
	}

	return nil
}

// UpdateFailedPaymentRequests update failed payment requests by given paymentRequestID, and notes.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) UpdateFailedPaymentRequests(ctx context.Context, paymentRequestID int64, notes string) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("id = ?", paymentRequestID).
		Updates(map[string]interface{}{
			"status":      "FAILED",
			"notes":       notes,
			"retry_count": gorm.Expr("retry_count + 1"),
			"update_time": time.Now(),
		}).Error
	if err != nil {
		return err
	}

	return nil
}

// UpdatePendingPaymentRequests update pending payment requests by given paymentRequestID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) UpdatePendingPaymentRequests(ctx context.Context, paymentRequestID int64) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("id = ?", paymentRequestID).
		Updates(map[string]interface{}{
			"status":      "PENDING",
			"update_time": time.Now(),
		}).Error
	if err != nil {
		return err
	}

	return nil
}

// GetExpiredPendingPayments get expired pending payments.
//
// It returns slice of models.Payment, and nil error when successful.
// Otherwise, nil value of models.Payment slice, and error will be returned.
func (r *paymentDatabase) GetExpiredPendingPayments(ctx context.Context) ([]models.Payment, error) {
	var result []models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("status = ? AND expired_time <= ?", "PENDING", time.Now()).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

// MarkExpired mark expired by given paymentID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) MarkExpired(ctx context.Context, paymentID int64) error {
	err := r.DB.Table("payments").WithContext(ctx).Model(&models.Payment{}).Where("id = ?", paymentID).
		Updates(map[string]interface{}{
			"status":      "EXPIRED",
			"update_time": time.Now(),
		}).Error
	if err != nil {
		return err
	}

	return nil
}

// InsertAuditLog insert audit log by given PaymentAuditLog.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *paymentDatabase) InsertAuditLog(ctx context.Context, param models.PaymentAuditLog) error {
	err := r.DB.Table("payment_audit_logs").WithContext(ctx).Create(param).Error
	if err != nil {
		return err
	}

	return nil
}

/*
1
2
3
4
5
6
7
8
..
...
1000
*/