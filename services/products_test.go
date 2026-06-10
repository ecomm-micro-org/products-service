package services

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ecomm-micro-org/products-service/cache"
	custom_errors "github.com/ecomm-micro-org/products-service/internal/constants/errors"
	"github.com/ecomm-micro-org/products-service/models"
	"github.com/ecomm-micro-org/products-service/pb"
	"github.com/ecomm-micro-org/products-service/store"
	"github.com/redis/go-redis/v9"
	"google.golang.org/genai"
)

func setupRedis(t *testing.T) *miniredis.Miniredis {
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
	return mini
}

func newTestProductService(t *testing.T) (*ProductService, *store.MemoryStore) {
	t.Helper()
	setupRedis(t)

	memStore := store.NewMemoryStore()
	svc := NewProductService(memStore, nil)
	svc.embedder = func(context.Context, *genai.Client, any) ([]float32, error) {
		return []float32{0.1, 0.2, 0.3}, nil
	}
	return svc, memStore
}

func sellerContext() context.Context {
	return context.WithValue(context.Background(), "role", "seller")
}

func TestProductServiceGetProductByIDCachesResult(t *testing.T) {
	svc, memStore := newTestProductService(t)
	product := &models.Product{ID: 1, Name: "Keyboard", Price: 99.99, Stock: 5, InStock: true}
	if err := memStore.AddProduct(product); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	first, err := svc.GetProductByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetProductByID() first call error = %v", err)
	}
	if first.Name != "Keyboard" {
		t.Fatalf("unexpected first response: %+v", first)
	}

	if err := memStore.UpdateProduct(&models.Product{ID: 1, Name: "Updated Keyboard", Price: 120, Stock: 5, InStock: true}); err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}

	second, err := svc.GetProductByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetProductByID() second call error = %v", err)
	}
	if second.Name != "Keyboard" {
		t.Fatalf("expected cached product name, got %q", second.Name)
	}
}

func TestProductServiceGetProductsByIDs(t *testing.T) {
	svc, memStore := newTestProductService(t)
	_ = memStore.AddProduct(&models.Product{ID: 1, Name: "Mouse", Price: 10})
	_ = memStore.AddProduct(&models.Product{ID: 2, Name: "Monitor", Price: 20})

	res, err := svc.GetProductsByIDs(context.Background(), []uint64{2, 1})
	if err != nil {
		t.Fatalf("GetProductsByIDs() error = %v", err)
	}
	if len(res.Products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(res.Products))
	}
	if res.Products[0].Id != 2 || res.Products[1].Id != 1 {
		t.Fatalf("unexpected product order: got %d then %d", res.Products[0].Id, res.Products[1].Id)
	}
}

func TestProductServiceCalculateTotalPrice(t *testing.T) {
	svc, memStore := newTestProductService(t)
	_ = memStore.AddProduct(&models.Product{ID: 1, Name: "Mouse", Price: 10})
	_ = memStore.AddProduct(&models.Product{ID: 2, Name: "Monitor", Price: 20})

	res, err := svc.CalculateTotalPrice([]*pb.OrderItem{
		{ProductId: 1, Quantity: 2},
		{ProductId: 2, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("CalculateTotalPrice() error = %v", err)
	}
	if res.TotalPrice != 40 {
		t.Fatalf("expected total price 40, got %v", res.TotalPrice)
	}
}

func TestProductServiceCalculateTotalPriceMissingProduct(t *testing.T) {
	svc, _ := newTestProductService(t)

	if _, err := svc.CalculateTotalPrice([]*pb.OrderItem{{ProductId: 99, Quantity: 1}}); err == nil {
		t.Fatal("expected missing product error")
	}
}

func TestProductServiceAddProduct(t *testing.T) {
	svc, memStore := newTestProductService(t)
	ctx := sellerContext()

	res, err := svc.AddProduct(ctx, &pb.AddProductRequest{
		Name:          "Desk Lamp",
		OriginalPrice: 25.5,
		Image:         "lamp.png",
		Category:      "Home",
		Description:   "Warm desk lamp",
		Stock:         8,
		InStock:       true,
		Tags:          []string{"lamp", "home"},
	})
	if err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}
	if res.Id == 0 {
		t.Fatal("expected AddProduct to assign an ID")
	}

	stored := models.NewProduct()
	if err := memStore.GetProductByID(res.Id, stored); err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if stored.Name != "Desk Lamp" {
		t.Fatalf("expected stored product name %q, got %q", "Desk Lamp", stored.Name)
	}
	if len(cache.Client().Keys(context.Background(), "product:*").Val()) != 1 {
		t.Fatal("expected one cached product entry")
	}
}

func TestProductServiceAddProductRequiresSellerRole(t *testing.T) {
	svc, _ := newTestProductService(t)

	if _, err := svc.AddProduct(context.Background(), &pb.AddProductRequest{Name: "Desk Lamp"}); err == nil {
		t.Fatal("expected seller role error")
	}
}

func TestProductServiceUpdateProduct(t *testing.T) {
	svc, memStore := newTestProductService(t)
	product := &models.Product{ID: 1, Name: "Chair", Price: 15, Stock: 5, InStock: true}
	if err := memStore.AddProduct(product); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	err := svc.UpdateProduct(sellerContext(), &pb.UpdateProductRequest{
		Id:            1,
		Name:          "Chair Pro",
		Price:         18,
		OriginalPrice: 20,
		Category:      "Furniture",
		Description:   "Updated",
		Stock:         4,
		InStock:       true,
		Tags:          []string{"seat"},
	})
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}

	stored := models.NewProduct()
	if err := memStore.GetProductByID(1, stored); err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if stored.Name != "Chair Pro" || stored.Price != 18 {
		t.Fatalf("unexpected updated product: %+v", stored)
	}
}

func TestProductServiceUpdateProductRequiresSellerRole(t *testing.T) {
	svc, _ := newTestProductService(t)

	err := svc.UpdateProduct(context.Background(), &pb.UpdateProductRequest{Id: 1})
	if err != custom_errors.ErrNotEnoughPermissions {
		t.Fatalf("expected ErrNotEnoughPermissions, got %v", err)
	}
}

func TestProductServiceDecreaseProductStock(t *testing.T) {
	svc, memStore := newTestProductService(t)
	if err := memStore.AddProduct(&models.Product{ID: 1, Name: "Cable", Stock: 2, InStock: true}); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	if err := svc.DecreaseProductStock(context.Background(), 1, 2); err != nil {
		t.Fatalf("DecreaseProductStock() error = %v", err)
	}

	stored := models.NewProduct()
	if err := memStore.GetProductByID(1, stored); err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if stored.Stock != 0 || stored.InStock {
		t.Fatalf("unexpected stock state: %+v", stored)
	}
}

func TestProductServiceDeleteProduct(t *testing.T) {
	svc, memStore := newTestProductService(t)
	if err := memStore.AddProduct(&models.Product{ID: 1, Name: "Cable", Stock: 2, InStock: true}); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	ctx := context.Background()
	cacheKey := "product:1"
	if err := cache.Client().Set(ctx, cacheKey, "cached", time.Minute).Err(); err != nil {
		t.Fatalf("cache.Set() error = %v", err)
	}

	if err := svc.DeleteProduct(sellerContext(), 1); err != nil {
		t.Fatalf("DeleteProduct() error = %v", err)
	}
	if err := memStore.GetProductByID(1, models.NewProduct()); err == nil {
		t.Fatal("expected product to be deleted")
	}
	if cache.Client().Exists(ctx, cacheKey).Val() != 0 {
		t.Fatal("expected cache entry to be deleted")
	}
}

func TestProductServiceDeleteProductRequiresSellerRole(t *testing.T) {
	svc, _ := newTestProductService(t)

	err := svc.DeleteProduct(context.Background(), 1)
	if err != custom_errors.ErrNotEnoughPermissions {
		t.Fatalf("expected ErrNotEnoughPermissions, got %v", err)
	}
}
