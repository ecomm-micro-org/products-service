package store

import (
	"encoding/json"
	"testing"

	"github.com/ecomm-micro-org/products-service/models"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

func TestMemoryStoreAddAndGetProductByID(t *testing.T) {
	s := NewMemoryStore()
	p := &models.Product{Name: "Keyboard", Price: 49.99, Tags: []string{"input"}}

	if err := s.AddProduct(p); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}
	if p.ID == 0 {
		t.Fatal("expected AddProduct to assign an ID")
	}

	got := models.NewProduct()
	if err := s.GetProductByID(p.ID, got); err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if got.Name != p.Name || got.Price != p.Price {
		t.Fatalf("GetProductByID() got %+v, want %+v", got, p)
	}

	got.Tags[0] = "changed"
	stored := models.NewProduct()
	if err := s.GetProductByID(p.ID, stored); err != nil {
		t.Fatalf("GetProductByID() second read error = %v", err)
	}
	if stored.Tags[0] != "input" {
		t.Fatalf("expected stored tags to be isolated, got %v", stored.Tags)
	}
}

func TestMemoryStoreGetProductByIDNotFound(t *testing.T) {
	s := NewMemoryStore()

	err := s.GetProductByID(99, models.NewProduct())
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestMemoryStoreGetProductsByIDs(t *testing.T) {
	s := NewMemoryStore()
	p1 := &models.Product{Name: "Mouse", Price: 10}
	p2 := &models.Product{Name: "Monitor", Price: 20}

	if err := s.AddProduct(p1); err != nil {
		t.Fatalf("AddProduct(p1) error = %v", err)
	}
	if err := s.AddProduct(p2); err != nil {
		t.Fatalf("AddProduct(p2) error = %v", err)
	}

	got, err := s.GetProductsByIDs([]uint64{p2.ID, 999, p1.ID})
	if err != nil {
		t.Fatalf("GetProductsByIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 products, got %d", len(got))
	}
	if got[0].ID != p2.ID || got[1].ID != p1.ID {
		t.Fatalf("unexpected order: got IDs %d, %d", got[0].ID, got[1].ID)
	}
}

func TestMemoryStoreUpdateProduct(t *testing.T) {
	s := NewMemoryStore()
	p := &models.Product{Name: "Headphones", Price: 75}
	if err := s.AddProduct(p); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	updated := &models.Product{ID: p.ID, Name: "Headphones Pro", Price: 99, InStock: true}
	if err := s.UpdateProduct(updated); err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}

	got := models.NewProduct()
	if err := s.GetProductByID(p.ID, got); err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if got.Name != "Headphones Pro" || got.Price != 99 {
		t.Fatalf("unexpected updated product: %+v", got)
	}
}

func TestMemoryStoreUpdateProductNotFound(t *testing.T) {
	s := NewMemoryStore()

	err := s.UpdateProduct(&models.Product{ID: 42, Name: "Ghost"})
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestMemoryStoreDecreaseProductStock(t *testing.T) {
	s := NewMemoryStore()
	p := &models.Product{Name: "SSD", Stock: 3, InStock: true}
	if err := s.AddProduct(p); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	updated, err := s.DecreaseProductStock(p.ID, 3)
	if err != nil {
		t.Fatalf("DecreaseProductStock() error = %v", err)
	}
	if updated.Stock != 0 || updated.InStock {
		t.Fatalf("expected stock 0 and inStock=false, got %+v", updated)
	}
}

func TestMemoryStoreDecreaseProductStockErrors(t *testing.T) {
	s := NewMemoryStore()
	p := &models.Product{Name: "SSD", Stock: 2, InStock: true}
	if err := s.AddProduct(p); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	if _, err := s.DecreaseProductStock(404, 1); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}

	if _, err := s.DecreaseProductStock(p.ID, 3); err == nil {
		t.Fatal("expected insufficient stock error")
	}
}

func TestMemoryStoreDeleteProduct(t *testing.T) {
	s := NewMemoryStore()
	p := &models.Product{Name: "Laptop"}
	if err := s.AddProduct(p); err != nil {
		t.Fatalf("AddProduct() error = %v", err)
	}

	if err := s.DeleteProduct(p.ID); err != nil {
		t.Fatalf("DeleteProduct() error = %v", err)
	}
	if err := s.GetProductByID(p.ID, models.NewProduct()); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected deleted product to be missing, got %v", err)
	}
}

func TestMemoryStoreDeleteProductNotFound(t *testing.T) {
	s := NewMemoryStore()

	if err := s.DeleteProduct(1); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestMemoryStoreAddEmbeddingAndCollectionLookup(t *testing.T) {
	s := NewMemoryStore()
	metadata, err := json.Marshal(map[string]any{"productId": 1})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	e := &models.Embedding{
		ID:           uuid.New(),
		CollectionID: "products",
		Embedding:    pgvector.NewVector([]float32{1, 2, 3}),
		Document:     "doc",
		Cmetadata:    metadata,
	}

	if err := s.AddEmbedding(e); err != nil {
		t.Fatalf("AddEmbedding() error = %v", err)
	}
	if len(s.embeddings) != 1 {
		t.Fatalf("expected one embedding to be stored, got %d", len(s.embeddings))
	}

	collectionID, err := s.GetCollectionByName("products")
	if err != nil {
		t.Fatalf("GetCollectionByName() error = %v", err)
	}
	if collectionID != "products" {
		t.Fatalf("expected collection ID 'products', got %q", collectionID)
	}
}

func TestMemoryStoreGetCollectionByNameNotFound(t *testing.T) {
	s := NewMemoryStore()

	if _, err := s.GetCollectionByName("missing"); err == nil {
		t.Fatal("expected collection lookup to fail")
	}
}
