package store

import (
	"products/models"
)

type MemoryStore struct {
	db map[uint]*models.Product
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		db: make(map[uint]*models.Product),
	}
}

func (s *MemoryStore) GetProductByID(id uint, p *models.Product) error {
	return nil
}

func (s *MemoryStore) GetProductsByIDs(productIDs []uint) ([]*models.Product, error) {
	var products []*models.Product

	return products, nil
}

func (s *MemoryStore) SearchProductsByKeyword(keyword string) ([]*models.Product, error) {
	var products []*models.Product

	return products, nil
}

func (s *MemoryStore) AddProduct(p *models.Product) error {
	return nil
}

func (s *MemoryStore) UpdateProduct(p *models.Product) error {
	return nil
}

func (s *MemoryStore) DecreaseProductStock(id uint, count uint) (*models.Product, error) {
	p := models.NewProduct()
	return p, nil
}

func (s *MemoryStore) DeleteProduct(id uint) error {
	return nil
}

func (s *MemoryStore) AddEmbedding(e *models.Embedding) error {
	return nil
}
