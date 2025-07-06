package repository

import (
	// golang package
	"context"
	"encoding/json"
	"fmt"
	"paymentfc/models"

	// external package
	"github.com/segmentio/kafka-go"
)

type PaymentEventPublisher interface {
	// PublishEventPaymentStatus publish event payment status by given orderID, status, and topic.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	PublishEventPaymentStatus(ctx context.Context, orderID int64, status string, topic string) error

	// PublishPaymentSuccess publish payment success by given orderID.
	//
	// It returns nil error when successful.
	// Otherwise, error will be returned.
	PublishPaymentSuccess(ctx context.Context, orderID int64) error
}

type kafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher new kafka publisher by given writer pointer of kafka.Writer.
//
// It returns PaymentEventPublisher when successful.
// Otherwise, empty PaymentEventPublisher will be returned.
func NewKafkaPublisher(writer *kafka.Writer) PaymentEventPublisher {
	return &kafkaPublisher{
		writer: writer,
	}
}

// PublishEventPaymentStatus publish event payment status by given orderID, status, and topic.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (k *kafkaPublisher) PublishEventPaymentStatus(ctx context.Context, orderID int64, status string, topic string) error {
	payload := models.PaymentStatusUpdateEvent{
		OrderID: orderID,
		Status:  status,
	}

	data, _ := json.Marshal(payload)
	return k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Topic: topic,
		Value: data,
	})
}

// PublishPaymentSuccess publish payment success by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (k *kafkaPublisher) PublishPaymentSuccess(ctx context.Context, orderID int64) error {
	payload := map[string]interface{}{
		"order_id": orderID,
		"status":   "paid",
	}

	data, _ := json.Marshal(payload)
	return k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Value: data,
	})
}

// PublishPaymentFailed publish payment failed by given orderID.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (k *kafkaPublisher) PublishPaymentFailed(ctx context.Context, orderID int64) error {
	payload := map[string]interface{}{
		"order_id": orderID,
		"status":   "failed",
	}

	data, _ := json.Marshal(payload)
	return k.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", orderID)),
		Value: data,
	})
}