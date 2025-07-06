package service

import (
	// golang package
	"context"
	"orderfc/cmd/order/repository"
	"orderfc/models"

	// external package
	"gorm.io/gorm"
)

type OrderService struct {
	OrderRepository repository.OrderRepository
}

// NewOrderService new orderservice by given OrderRepository.
//
// It returns pointer of OrderService when successful.
// Otherwise, nil pointer of OrderService will be returned.
func NewOrderService(orderRepository repository.OrderRepository) *OrderService {
	return &OrderService{
		OrderRepository: orderRepository,
	}
}

// CheckIdempotency check idempotency by given token.
//
// It returns bool, and nil error when successful.
// Otherwise, empty bool, and error will be returned.
func (s *OrderService) CheckIdempotency(ctx context.Context, token string) (bool, error) {
	isExist, err := s.OrderRepository.CheckIdempotency(ctx, token)
	if err != nil {
		return false, err
	}

	return isExist, nil
}

// SaveIdempotencyToken save idempotency token by given token.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (s *OrderService) SaveIdempotencyToken(ctx context.Context, token string) error {
	err := s.OrderRepository.SaveIdempotencyToken(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

// GetOriderInfoByOrderID get orider info by order id by given orderID.
//
// It returns models.Order, and nil error when successful.
// Otherwise, empty models.Order, and error will be returned.
func (s *OrderService) GetOrderInfoByOrderID(ctx context.Context, orderID int64) (models.Order, error) {
	orderInfo, err := s.OrderRepository.GetOrderInfoByOrderID(ctx, orderID)
	if err != nil {
		return models.Order{}, err
	}

	return orderInfo, nil
}

// GetOrderDetailByOrderDetailID get order detail by order detail id by given orderDetailID.
//
// It returns models.OrderDetail, and nil error when successful.
// Otherwise, empty models.OrderDetail, and error will be returned.
func (s *OrderService) GetOrderDetailByOrderDetailID(ctx context.Context, orderDetailID int64) (models.OrderDetail, error) {
	orderDetail, err := s.OrderRepository.GetOrderDetailByOrderDetailID(ctx, orderDetailID)
	if err != nil {
		return models.OrderDetail{}, err
	}

	return orderDetail, nil
}

// UpdateOrderStatus update order status by given orderID, and status.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status int) error {
	err := s.OrderRepository.UpdateOrderStatus(ctx, orderID, status)
	if err != nil {
		return err
	}

	return nil
}

// SaveOrderWithDetail save order with detail by given order pointer of models.Order, and detail pointer of models.OrderDetail.
//
// It returns int64, and nil error when successful.
// Otherwise, empty int64, and error will be returned.
func (s *OrderService) SaveOrderAndOrderDetail(ctx context.Context, order *models.Order, orderDetail *models.OrderDetail) (int64, error) {
	var orderID int64

	err := s.OrderRepository.WithTransaction(ctx, func(tx *gorm.DB) error {
		err := s.OrderRepository.InsertOrderDetailTx(ctx, tx, orderDetail)
		if err != nil {
			return err
		}

		order.OrderDetailID = orderDetail.ID

		err = s.OrderRepository.InsertOrderTx(ctx, tx, order)
		if err != nil {
			return err
		}

		orderID = order.ID
		return nil
	})

	if err != nil {
		return 0, err
	}

	return orderID, nil
}

// GetOrderHistoriesByUserID get order histories by user id by given userID.
//
// It returns slice of models.OrderHistoryResponse, and nil error when successful.
// Otherwise, nil value of models.OrderHistoryResponse slice, and error will be returned.
func (s *OrderService) GetOrderHistoriesByUserID(ctx context.Context, param models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	orderHistories, err := s.OrderRepository.GetOrderHistoriesByUserID(ctx, param)
	if err != nil {
		return nil, err
	}
	return orderHistories, nil
}