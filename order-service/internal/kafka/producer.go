package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"

	"order-service/internal/models"
)

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafkago.Hash{}, // key = order_id -> cùng order vào cùng partition
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

// PublishOrderCreated publish fire-and-forget: lỗi chỉ log, không chặn response cho Frontend
// (đúng luồng async đã ghi trong system-architecture.md mục 4).
func (p *Producer) PublishOrderCreated(ctx context.Context, event models.OrderCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal order-created event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(event.OrderID.String()),
		Value: payload,
	})
}
