package handlers

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/ecomm-micro-org/products-service/cache"
	"github.com/ecomm-micro-org/products-service/models"
	"github.com/ecomm-micro-org/products-service/pb"
	"github.com/ecomm-micro-org/products-service/services"
	"github.com/ecomm-micro-org/products-service/store"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupHandlerService(t *testing.T) (*ProductHandler, *store.MemoryStore) {
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
	svc := services.NewProductService(memStore, nil)
	return NewProductHandler(svc), memStore
}

func TestHandlerGetProductByIDNotFound(t *testing.T) {
	h, _ := setupHandlerService(t)

	_, err := h.GetProductByID(context.Background(), &pb.GetProductByIDRequest{Id: 999})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestHandlerGetProductByIDSuccess(t *testing.T) {
	h, memStore := setupHandlerService(t)
	if err := memStore.AddProduct(&models.Product{ID: 1, Name: "Keyboard", Price: 10}); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	res, err := h.GetProductByID(context.Background(), &pb.GetProductByIDRequest{Id: 1})
	if err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if res.Id != 1 || res.Name != "Keyboard" {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestHandlerGetProductsByIDsStreamsProducts(t *testing.T) {
	h, memStore := setupHandlerService(t)
	_ = memStore.AddProduct(&models.Product{ID: 1, Name: "Mouse", Price: 10})
	_ = memStore.AddProduct(&models.Product{ID: 2, Name: "Monitor", Price: 20})

	res, err := h.GetProductsByIDs(context.Background(), &pb.GetProductsByIDsRequest{ProductIds: []uint64{2, 1}})
	if err != nil {
		t.Fatalf("GetProductsByIDs() error = %v", err)
	}
	if len(res.Products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(res.Products))
	}
}

func TestHandlerCalculateTotalPrice(t *testing.T) {
	h, memStore := setupHandlerService(t)
	_ = memStore.AddProduct(&models.Product{ID: 1, Name: "Mouse", Price: 10})
	_ = memStore.AddProduct(&models.Product{ID: 2, Name: "Monitor", Price: 20})

	res, err := h.CalculateTotalPrice(context.Background(), &pb.CalculateTotalPriceRequest{
		OrderItems: []*pb.OrderItem{{ProductId: 1, Quantity: 2}, {ProductId: 2, Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("CalculateTotalPrice() error = %v", err)
	}
	if res.TotalPrice != 40 {
		t.Fatalf("expected total price 40, got %v", res.TotalPrice)
	}
}

func TestHandlerUpdateProductPermissionDenied(t *testing.T) {
	h, _ := setupHandlerService(t)

	_, err := h.UpdateProduct(context.Background(), &pb.UpdateProductRequest{Id: 1})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", status.Code(err))
	}
}

func TestHandlerDeleteProductPermissionDenied(t *testing.T) {
	h, _ := setupHandlerService(t)

	_, err := h.DeleteProduct(context.Background(), &pb.DeleteProductRequest{Id: 1})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", status.Code(err))
	}
}
