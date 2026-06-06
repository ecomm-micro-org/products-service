package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ecomm-micro-org/products-service/services"
	"github.com/segmentio/kafka-go"
)

var (
	once sync.Once
)

type items []struct {
	ID       uint64 `json:"product_id"`
	Quantity uint64 `json:"quantity"`
}

type consumer struct {
	t              Topic
	r              *kafka.Reader
	ps             *services.ProductService
	processingChan chan *items
	kafkaErr       chan<- error
}

func NewConsumer(cCfg ConsumerConfig, ps *services.ProductService, kafkaErr chan<- error) *consumer {
	return &consumer{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cCfg.Brokers,
			Topic:          cCfg.Topic,
			GroupID:        cCfg.GroupID,
			MinBytes:       cCfg.MinBytes,
			MaxBytes:       cCfg.MaxBytes,
			CommitInterval: time.Duration(cCfg.CommitInterval) * time.Millisecond,
			StartOffset:    cCfg.StartOffset,
		}),
		ps:             ps,
		processingChan: make(chan *items),
		kafkaErr:       kafkaErr,
	}
}

func (c *consumer) Init() error {
	if c.processingChan == nil || c.kafkaErr == nil {
		return fmt.Errorf("processing and error channels cannot be nil")
	}
	once.Do(func() {
		go c.processData()
	})
	return nil
}

func (c *consumer) Consume() {
	ctx := context.Background()
	for {
		msg, err := c.r.FetchMessage(ctx)
		if err != nil {
			c.kafkaErr <- err
			continue
		}

		data := new(items)
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			c.kafkaErr <- err
			continue
		}

		c.processingChan <- data

		if err := c.r.CommitMessages(ctx, msg); err != nil {
			c.kafkaErr <- err
			continue
		}
	}
}

func (c *consumer) processData() {
	for orderedItems := range c.processingChan {
		go func(oi *items) {
			for _, i := range *oi {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				err := c.ps.DecreaseProductStock(ctx, i.ID, i.Quantity)
				cancel()

				if err != nil {
					c.kafkaErr <- err
					continue
				}
			}
		}(orderedItems)
	}
}

func (c *consumer) Close() error {
	close(c.processingChan)
	return c.r.Close()
}
