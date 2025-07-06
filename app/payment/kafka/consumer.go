package kafka

import (
	// golang package
	"context"
	"encoding/json"
	"log"
	"paymentfc/models"

	// external package
	"github.com/segmentio/kafka-go"
)

// StartOrderConsumer start order consumer by given broker, topic, and handler.
func StartOrderConsumer(broker string, topic string, handler func(models.OrderCreatedEvent)) {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "paymentfc",
	})

	go func(r *kafka.Reader) {
		for {
			message, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Println("Error Read Message Kafka: ", err.Error())
				// to do improvement: store data to DB
				continue
			}

			var event models.OrderCreatedEvent
			err = json.Unmarshal(message.Value, &event)
			if err != nil {
				log.Println("Error Unmarshal Message: ", err.Error())
				continue
			}

			log.Printf("Received Event Order Created: %+v", event)
			handler(event)
		}
	}(consumer)
}