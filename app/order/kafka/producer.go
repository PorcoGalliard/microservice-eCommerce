package kafka

import (
	// golang package
	"context"
	"encoding/json"
	"fmt"
	"orderfc/models"

	// external package
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer new kafka producer by given slice of brokers, and topic.
//
// It returns pointer of KafkaProducer when successful.
// Otherwise, nil pointer of KafkaProducer will be returned.
func NewKafkaProducer(brokers []string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaProducer{writer: writer}
}

// PublishOrderCreated publish order created by given event.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (p *KafkaProducer) PublishOrderCreated(ctx context.Context, event models.OrderCreatedEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", event.OrderID)),
		Value: value,
		Topic: "order.created",
	}

	return p.writer.WriteMessages(ctx, msg)
}

// PublishProductStockUpdate publish product stock update by given ProductStockUpdateEvent.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (p *KafkaProducer) PublishProductStockUpdate(ctx context.Context, event models.ProductStockUpdateEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", event.OrderID)),
		Value: value,
		Topic: "stock.update",
	}

	return p.writer.WriteMessages(ctx, msg)
}

// PublishProductStockRollback publish product stock rollback by given ProductStockUpdateEvent.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (p *KafkaProducer) PublishProductStockRollback(ctx context.Context, event models.ProductStockUpdateEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("order-%d", event.OrderID)),
		Value: value,
		Topic: "stock.rollback",
	}

	return p.writer.WriteMessages(ctx, msg)
}

// Close close.
//
// It returns nil error when successful.
// Otherwise, error will be returned.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// iphone stock 2
// user A --> checkout iphone qty 1
// user B --> checkout iphone qty 2 --> error
// product perlu tahu ketika ada order.created --> temp menjaga stock sebelum diupdate permanen