package usecase

import (
	// golang package
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"orderfc/cmd/order/service"
	"orderfc/infrastructure/constant"
	"orderfc/kafka"
	"orderfc/models"
	"time"
)

type OrderUsecase struct {
	OrderService service.OrderService
	Producer     kafka.KafkaProducer
}

// NewOrderUsecase new orderusecase by given OrderService.
//
// It returns pointer of OrderUsecase when successful.
// Otherwise, nil pointer of OrderUsecase will be returned.
func NewOrderUsecase(orderService service.OrderService, kafkaProducer kafka.KafkaProducer) *OrderUsecase {
	return &OrderUsecase{
		OrderService: orderService,
		Producer:     kafkaProducer,
	}
}

// CheckoutOrder checkout order by given CheckoutRequest.
//
// It returns int64, and nil error when successful.
// Otherwise, empty int64, and error will be returned.
func (uc *OrderUsecase) CheckoutOrder(ctx context.Context, param *models.CheckoutRequest) (int64, error) {
	if param.IdempotencyToken != "" {
		isExist, err := uc.OrderService.CheckIdempotency(ctx, param.IdempotencyToken)
		if err != nil {
			return 0, err
		}

		if isExist {
			return 0, errors.New("order already created, please check again!")
		}
	}

	if err := uc.validateProducts(param.Items); err != nil {
		return 0, err
	}

	totalQty, totalAmount := uc.calculateOrderSummary(param.Items)
	productJSON, historyJSON, err := uc.constructOrderDetail(param.Items)
	if err != nil {
		return 0, err
	}

	orderDetail := &models.OrderDetail{
		Products:     productJSON,
		OrderHistory: historyJSON,
	}

	order := &models.Order{
		UserID:          param.UserID,
		Amount:          totalAmount,
		TotalQty:        totalQty,
		Status:          constant.OrderStatusCreated,
		PaymentMethod:   param.PaymentMethod,
		ShippingAddress: param.ShippingAddress,
	}

	orderID, err := uc.OrderService.SaveOrderAndOrderDetail(ctx, order, orderDetail)
	if err != nil {
		return 0, err
	}

	if param.IdempotencyToken != "" {
		_ = uc.OrderService.SaveIdempotencyToken(ctx, param.IdempotencyToken)
	}

	event := models.OrderCreatedEvent{
		OrderID:         orderID,
		UserID:          param.UserID,
		TotalAmount:     order.Amount,
		PaymentMethod:   param.PaymentMethod,
		ShippingAddress: param.ShippingAddress,
	}

	go func() {
		if err := uc.Producer.PublishOrderCreated(context.Background(), event); err != nil {
			fmt.Println("Failed to send Kafka event:", err)
		} else {
			fmt.Println("OrderCreated event sent to Kafka")
		}
	}()

	updateStockEvent := models.ProductStockUpdateEvent{
		OrderID:   orderID,
		Products:  convertCheckoutItemToProductItems(param.Items),
		EventTime: time.Now(),
	}

	go func() {
		if err := uc.Producer.PublishProductStockUpdate(context.Background(), updateStockEvent); err != nil {
			fmt.Println("Failed to send kafka event: ", err)
		} else {
			fmt.Println("StockUpdate event sent to kafka")
		}
	}()

	return orderID, nil
}

// validateProducts validate products by given items slice of CheckoutItem.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (uc *OrderUsecase) validateProducts(items []models.CheckoutItem) error {
	seen := map[int64]bool{}
	for _, item := range items {
		if seen[item.ProductID] {
			return fmt.Errorf("duplicate product: %d", item.ProductID)
		}
		seen[item.ProductID] = true

		if item.Quantity <= 0 || item.Quantity > 10000 {
			return fmt.Errorf("invalid quantity for %d", item.ProductID)
		}

		if item.Price <= 0 {
			return fmt.Errorf("invalid price for %d", item.ProductID)
		}
	}
	return nil
}

// calculateOrderSummary calculate order summary by given items slice of CheckoutItem.
//
// It returns int, and float64 when successful.
// Otherwise, empty int, and empty float64 will be returned.
func (uc *OrderUsecase) calculateOrderSummary(items []models.CheckoutItem) (int, float64) {
	var totalQty int
	var totalAmount float64
	for _, item := range items {
		totalQty += item.Quantity
		totalAmount += float64(item.Quantity) * item.Price
	}
	return totalQty, totalAmount
}

// constructOrderDetail construct order detail by given items slice of CheckoutItem.
//
// It returns string, string, and nil error when successful.
// Otherwise, empty string, empty string, and error will be returned.
func (uc *OrderUsecase) constructOrderDetail(items []models.CheckoutItem) (string, string, error) {
	productsJSON, _ := json.Marshal(items)
	history := []map[string]interface{}{
		{"status": "created", "timestamp": time.Now()},
	}
	historyJSON, _ := json.Marshal(history)

	return string(productsJSON), string(historyJSON), nil
}

// GetOrderHistory get order history by given userID.
//
// It returns slice of models.OrderHistoryResponse, and nil error when successful.
// Otherwise, nil value of models.OrderHistoryResponse slice, and error will be returned.
func (uc *OrderUsecase) GetOrderHistory(ctx context.Context, param models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	orderHistory, err := uc.OrderService.GetOrderHistoriesByUserID(ctx, param)
	if err != nil {
		return nil, err
	}

	return orderHistory, nil
}

// convertCheckoutItemToProductItems convert checkout item to product items by given source slice of CheckoutItem.
//
// It returns slice of models.ProductItem when successful.
// Otherwise, nil value of models.ProductItem slice will be returned.
func convertCheckoutItemToProductItems(source []models.CheckoutItem) []models.ProductItem {
	result := make([]models.ProductItem, len(source))
	for index, item := range source {
		result[index] = models.ProductItem{
			ProductID: item.ProductID,
			Qty:       item.Quantity,
		}
	}

	return result
}