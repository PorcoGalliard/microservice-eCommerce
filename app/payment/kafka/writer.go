package kafka

import (
	"github.com/segmentio/kafka-go"
)

// NewWriter new writer by given broker, and topic.
//
// It returns pointer of kafka.Writer when successful.
// Otherwise, nil pointer of kafka.Writer will be returned.
func NewWriter(broker string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}