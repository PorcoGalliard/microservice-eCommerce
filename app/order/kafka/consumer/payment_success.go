package consumer

import (
	// golang package
	"context"
	"encoding/json"
	"orderfc/cmd/order/service"
	"orderfc/infrastructure/constant"
	"orderfc/infrastructure/log"
	kafkaFC "orderfc/kafka"
	"orderfc/models"

	// external package
	"github.com/segmentio/kafka-go"
)

type PaymentSuccessConsumer struct {
	Reader       *kafka.Reader
	Producer     kafkaFC.KafkaProducer
	OrderService service.OrderService
}

// NewPaymentSuccessConsumer new payment success consumer by given slice of brokers, topic, and OrderService.
//
// It returns pointer of PaymentSuccessConsumer when successful.
// Otherwise, nil pointer of PaymentSuccessConsumer will be returned.
func NewPaymentSuccessConsumer(brokers []string, topic string, orderService service.OrderService, kafkaProducer kafkaFC.KafkaProducer) *PaymentSuccessConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "orderfc",
	})

	return &PaymentSuccessConsumer{
		Reader:       reader,
		OrderService: orderService,
		Producer:     kafkaProducer,
	}
}

// ConsumerPaymentSuccess consumer payment success.
func (c *PaymentSuccessConsumer) StartPaymentSuccessConsumer(ctx context.Context) {
	log.Logger.Println("[KAFKA] Listening to topic: payment.success")

	for {
		// order id: 1 success --> skipped --> high risk (P0)
		// order id: 2 success
		message, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Logger.Println("[KAFKA] Error Read Message: ", err)
			continue
		}

		var event models.PaymentUpdateStatusEvent
		err = json.Unmarshal(message.Value, &event)
		if err != nil {
			log.Logger.Println("[KAFKA] Error Unmarshal Event Message Value: ", err)
			continue
		}

		log.Logger.Printf("[KAFKA] Received payment.success event for Order ID #%d", event.OrderID)

		// update DB
		err = c.OrderService.UpdateOrderStatus(ctx, event.OrderID, constant.OrderStatusCompleted)
		if err != nil {
			log.Logger.Println("[KAFKA] Error Update Order Status: ", err)
			continue
		}

		// // get order info from db
		// orderInfo, err := c.OrderService.GetOrderInfoByOrderID(ctx, event.OrderID)
		// if err != nil {
		// 	log.Logger.Println("[KAFKA] Error Get Order Info: ", err)
		// 	continue
		// }

		// // get order detail from db
		// orderDetail, err := c.OrderService.GetOrderDetailByOrderDetailID(ctx, orderInfo.OrderDetailID)
		// if err != nil {
		// 	log.Logger.Println("[KAFKA] Error Get Order Detail: ", err)
		// 	continue
		// }

		// // get product list from order detail
		// var products []models.CheckoutItem
		// err = json.Unmarshal([]byte(orderDetail.Products), &products)
		// if err != nil {
		// 	log.Logger.Println("[KAFKA] Error Get Product List From Order Detail: ", err)
		// 	continue
		// }

		// publish event product service
		// err = c.Producer.PublishProductStockUpdate(ctx, models.ProductStockUpdateEvent{
		// 	OrderID:   event.OrderID,
		// 	Products:  convertCheckoutItemToProductItems(products),
		// 	EventTime: time.Now(),
		// })
	}
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