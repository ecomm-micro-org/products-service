package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ecomm-micro-org/products-service/cache"
	"github.com/ecomm-micro-org/products-service/models"
	"github.com/ecomm-micro-org/products-service/services"
	"github.com/ecomm-micro-org/products-service/store"
	"github.com/redis/go-redis/v9"
)

func setupKafkaTestService(t *testing.T) (*services.ProductService, *store.MemoryStore) {
	t.Helper()
	mini := miniredis.RunT(t)
	cache.SetClient(redis.NewClient(&redis.Options{Addr: mini.Addr()}))
	t.Cleanup(func() {
		if client := cache.Client(); client != nil {
			_ = client.Close()
		}
		cache.SetClient(nil)
		mini.Close()
	})

	memStore := store.NewMemoryStore()
	return services.NewProductService(memStore, nil), memStore
}

func TestConsumerInitValidatesChannels(t *testing.T) {
	c := &consumer{}
	if err := c.Init(); err == nil {
		t.Fatal("expected Init to fail when channels are nil")
	}
}

func TestConsumerProcessDataDecreasesStock(t *testing.T) {
	svc, memStore := setupKafkaTestService(t)
	if err := memStore.AddProduct(&models.Product{ID: 1, Name: "Keyboard", Stock: 3, InStock: true}); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	c := &consumer{
		ps:             svc,
		processingChan: make(chan *items),
		kafkaErr:       make(chan error, 1),
	}

	go c.processData()
	c.processingChan <- &items{{ID: 1, Quantity: 2}}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		stored := models.NewProduct()
		if err := memStore.GetProductByID(1, stored); err == nil && stored.Stock == 1 {
			close(c.processingChan)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	close(c.processingChan)

	stored := models.NewProduct()
	_ = memStore.GetProductByID(1, stored)
	t.Fatalf("expected stock to decrease to 1, got %+v", stored)
}

func TestConsumerProcessDataReportsErrors(t *testing.T) {
	svc, _ := setupKafkaTestService(t)
	errCh := make(chan error, 1)
	c := &consumer{
		ps:             svc,
		processingChan: make(chan *items),
		kafkaErr:       errCh,
	}

	go c.processData()
	c.processingChan <- &items{{ID: 99, Quantity: 1}}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected a non-nil error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for kafka error")
	}
	close(c.processingChan)

	if cache.Client().Exists(context.Background(), "product:99").Val() != 0 {
		t.Fatal("did not expect cache entry for missing product")
	}
}
