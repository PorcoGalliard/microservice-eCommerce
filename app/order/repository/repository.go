package repository

import (
	// golang package
	"context"
	"encoding/json"
	"errors"
	"orderfc/infrastructure/constant"
	"orderfc/models"
	"time"

	// external package
	"gorm.io/gorm"
)

// InsertOrderDetailTx insert order detail tx by given tx pointer of gorm.DB, and detail pointer of models.OrderDetail.
//
// It returns int64, and nil error when successful.
// Otherwise, empty int64, and error will be returned.
func (r *OrderRepository) InsertOrderDetailTx(ctx context.Context, tx *gorm.DB, orderDetail *models.OrderDetail) error {
	err := tx.WithContext(ctx).Table("order_detail").Create(orderDetail).Error
	return err
}

// InsertOrderTx insert order tx by given tx pointer of gorm.DB, and order pointer of models.Order.
//
// It returns int64, and nil error when successful.
// Otherwise, empty int64, and error will be returned.
func (r *OrderRepository) InsertOrderTx(ctx context.Context, tx *gorm.DB, order *models.Order) error {
	err := tx.WithContext(ctx).Table("orders").Create(order).Error
	return err
}

// WithTransaction with transaction by given fn.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *OrderRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := r.Database.Begin().WithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CheckIdempotency check idempotency by given token.
//
// It returns bool, and nil error when successful.
// Otherwise, empty bool, and error will be returned.
func (r *OrderRepository) CheckIdempotency(ctx context.Context, token string) (bool, error) {
	var log models.OrderRequestLog
	err := r.Database.WithContext(ctx).Table("order_request_log").First(&log, "idempotency_token = ?", token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

// SaveIdempotencyToken save idempotency token by given token.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *OrderRepository) SaveIdempotencyToken(ctx context.Context, token string) error {
	log := models.OrderRequestLog{
		IdempotencyToken: token,
		CreateTime:       time.Now(),
	}
	return r.Database.WithContext(ctx).Table("order_request_log").Create(&log).Error
}

// UpdateOrderStatus update order status by given orderID, and status.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status int) error {
	err := r.Database.Table("orders").WithContext(ctx).Model(&models.Order{}).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now(),
		}).Where("order_id = ?", orderID).Error

	if err != nil {
		return err
	}

	return nil
}

// GetOrderInfoByOrderID get order info by order id by given orderID.
//
// It returns models.Order, and nil error when successful.
// Otherwise, empty models.Order, and error will be returned.
func (r *OrderRepository) GetOrderInfoByOrderID(ctx context.Context, orderID int64) (models.Order, error) {
	var result models.Order
	err := r.Database.Table("orders").WithContext(ctx).Where("order_id = ?", orderID).Find(&result).Error
	if err != nil {
		return models.Order{}, err
	}

	return result, nil
}

// GetOrderDetailByOrderDetailID get order detail by order detail id by given orderDetailID.
//
// It returns models.OrderDetail, and nil error when successful.
// Otherwise, empty models.OrderDetail, and error will be returned.
func (r *OrderRepository) GetOrderDetailByOrderDetailID(ctx context.Context, orderDetailID int64) (models.OrderDetail, error) {
	var result models.OrderDetail
	err := r.Database.Table("order_detail").WithContext(ctx).Where("id = ?", orderDetailID).Find(&result).Error
	if err != nil {
		return models.OrderDetail{}, err
	}

	return result, nil
}

// GetOrderHistoriesByUserID get order histories by user id by given userID.
//
// It returns slice of models.OrderHistoryResponse, and nil error when successful.
// Otherwise, nil value of models.OrderHistoryResponse slice, and error will be returned.
func (r *OrderRepository) GetOrderHistoriesByUserID(ctx context.Context, param models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	var results []models.OrderJoinResult

	query := r.Database.WithContext(ctx).
		Table("orders AS o").
		Select(`o.id, o.amount, o.total_qty, o.status, o.payment_method, o.shipping_address,
	        d.products, d.order_history`).
		Joins("JOIN order_detail d ON o.order_detail_id = d.id").
		Where("o.user_id = ?", param.UserID)

	if param.Status > 0 {
		query = query.Where("o.status = ?", param.Status)
	}

	err := query.
		Order("o.id DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Transform to OrderHistoryResponse
	var response []models.OrderHistoryResponse
	for _, row := range results {
		var products []models.CheckoutItem
		var history []models.StatusHistory

		_ = json.Unmarshal([]byte(row.Products), &products)
		_ = json.Unmarshal([]byte(row.OrderHistory), &history)

		response = append(response, models.OrderHistoryResponse{
			OrderID:         row.ID,
			TotalAmount:     row.Amount,
			TotalQty:        row.TotalQty,
			Status:          constant.OrderStatusTranslated[row.Status],
			PaymentMethod:   row.PaymentMethod,
			ShippingAddress: row.ShippingAddress,
			Products:        products,
			History:         history,
		})
	}

	return response, nil
}