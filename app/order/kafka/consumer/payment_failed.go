package consumer

import (
	// golang package
	"context"
	"encoding/json"
	"log"
	"orderfc/cmd/order/service"
	"orderfc/infrastructure/constant"
	kafkaFC "orderfc/kafka"
	"orderfc/models"
	"time"

	// external package
	"github.com/segmentio/kafka-go"
)

type PaymentFailedConsumer struct {
	Reader       *kafka.Reader
	Producer     kafkaFC.KafkaProducer
	OrderService service.OrderService
}

// NewPaymentFailedConsumer new payment failed consumer by given slice of brokers, topic, OrderService, and KafkaProducer.
//
// It returns PaymentFailedConsumer when successful.
// Otherwise, empty PaymentFailedConsumer will be returned.
func NewPaymentFailedConsumer(brokers []string, topic string, orderService service.OrderService, kafkaProducer kafkaFC.KafkaProducer) *PaymentFailedConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "orderfc",
	})

	return &PaymentFailedConsumer{
		Reader:       reader,
		Producer:     kafkaProducer,
		OrderService: orderService,
	}
}

// Start start.
func (c *PaymentFailedConsumer) Start(ctx context.Context) {
	log.Println("Listening to topic payment.failed")

	for {
		message, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Println("[PF] Failed to ReadMessage: ", err)
			continue
		}

		var event models.PaymentUpdateStatusEvent
		err = json.Unmarshal(message.Value, &event)
		if err != nil {
			log.Println("[PF] Error Unmarshal Payment Update Status Event: ", err)
			continue
		}

		// update DB status order
		err = c.OrderService.UpdateOrderStatus(ctx, event.OrderID, constant.OrderStatusCancelled)
		if err != nil {
			log.Println("[PF] Error Update Order Status: ", err)
			continue
		}

		// order info
		orderInfo, err := c.OrderService.GetOrderInfoByOrderID(ctx, event.OrderID)
		if err != nil {
			continue
		}

		// order detail info
		orderDetailInfo, err := c.OrderService.GetOrderDetailByOrderDetailID(ctx, orderInfo.OrderDetailID)
		if err != nil {
			continue
		}

		// construct products
		var products []models.CheckoutItem
		err = json.Unmarshal([]byte(orderDetailInfo.Products), &products)
		if err != nil {
			continue
		}

		// publish event stock.rollback
		updateStockEvent := models.ProductStockUpdateEvent{
			OrderID:   event.OrderID,
			Products:  convertCheckoutItemToProductItems(products),
			EventTime: time.Now(),
		}

		err = c.Producer.PublishProductStockRollback(ctx, updateStockEvent)
		if err != nil {
			continue
		}
	}
}