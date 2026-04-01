package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"products/internal/config"
	"products/services"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	consumer *kafka.Reader
}

func NewConsumer(topic Topic) *Consumer {
	return &Consumer{
		consumer: kafka.NewReader(kafka.ReaderConfig{
			Brokers:           config.Config().Brokers,
			Topic:             topic.String(),
			GroupID:           "products-service",
			MinBytes:          1, // don't wait for data to accumulate
			MaxBytes:          10e6,
			MaxWait:           3 * time.Second,  // wait longer per fetch
			HeartbeatInterval: 3 * time.Second,  // heartbeat every 3s
			SessionTimeout:    30 * time.Second, // give 30s before eviction
			RebalanceTimeout:  30 * time.Second,
		}),
	}
}

func (c *Consumer) Consume(s *services.ProductService, errs chan<- error) {
	if c.consumer.Config().Topic != TopicOrderCreated.String() {
		errs <- fmt.Errorf("Consume requires topic to be set to orders.created topic")
		return
	}

	ctx := context.Background()
	for {
		msg, err := c.consumer.FetchMessage(ctx)
		if err != nil {
			errs <- err
			continue
		}

		var data []struct {
			ID       uint `json:"product_id"`
			Quantity uint `json:"quantity"`
		}
		err = json.Unmarshal(msg.Value, &data)
		if err != nil {
			errs <- err
			continue
		}

		go func(items []struct {
			ID       uint `json:"product_id"`
			Quantity uint `json:"quantity"`
		}) {
			for _, v := range items {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				err := s.DecreaseProductStock(ctx, v.ID, v.Quantity)
				cancel()
				if err != nil {
					errs <- err
					continue
				}
			}
		}(data)

		if err := c.consumer.CommitMessages(ctx, msg); err != nil {
			errs <- err
			continue
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
