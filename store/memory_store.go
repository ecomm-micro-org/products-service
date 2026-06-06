package store

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ecomm-micro-org/products-service/models"
	"gorm.io/gorm"
)

type MemoryStore struct {
	mu          sync.RWMutex
	db          map[uint64]*models.Product
	embeddings  map[string]*models.Embedding
	collections map[string]string
	nextID      uint64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		db:         make(map[uint64]*models.Product),
		embeddings: make(map[string]*models.Embedding),
		collections: map[string]string{
			"products": "products",
		},
		nextID: 1,
	}
}

func cloneProduct(p *models.Product) *models.Product {
	if p == nil {
		return nil
	}

	clone := *p
	if p.Tags != nil {
		clone.Tags = append([]string(nil), p.Tags...)
	}
	return &clone
}

func cloneEmbedding(e *models.Embedding) *models.Embedding {
	if e == nil {
		return nil
	}

	clone := *e
	if e.Cmetadata != nil {
		clone.Cmetadata = append([]byte(nil), e.Cmetadata...)
	}
	values := e.Embedding.Slice()
	if values != nil {
		clone.Embedding = e.Embedding
	}
	return &clone
}

func (s *MemoryStore) GetProductByID(id uint64, p *models.Product) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, ok := s.db[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}

	*p = *cloneProduct(product)
	return nil
}

func (s *MemoryStore) GetProductsByIDs(productIDs []uint64) ([]*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]*models.Product, 0, len(productIDs))
	for _, id := range productIDs {
		if product, ok := s.db[id]; ok {
			products = append(products, cloneProduct(product))
		}
	}

	return products, nil
}

func (s *MemoryStore) AddProduct(p *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.ID == 0 {
		p.ID = s.nextID
		s.nextID++
	} else {
		if _, ok := s.db[p.ID]; ok {
			return fmt.Errorf("product already exists")
		}
		if p.ID >= s.nextID {
			s.nextID = p.ID + 1
		}
	}

	s.db[p.ID] = cloneProduct(p)
	return nil
}

func (s *MemoryStore) UpdateProduct(p *models.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.db[p.ID]; !ok {
		return gorm.ErrRecordNotFound
	}

	s.db[p.ID] = cloneProduct(p)
	return nil
}

func (s *MemoryStore) DecreaseProductStock(id uint64, count uint64) (*models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, ok := s.db[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	if product.Stock < count {
		return nil, errors.New("unable to place order due to insufficient stock")
	}

	updated := cloneProduct(product)
	updated.Stock -= count
	if updated.Stock == 0 {
		updated.InStock = false
	}

	s.db[id] = updated
	return cloneProduct(updated), nil
}

func (s *MemoryStore) DeleteProduct(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.db[id]; !ok {
		return gorm.ErrRecordNotFound
	}

	delete(s.db, id)
	return nil
}

func (s *MemoryStore) AddEmbedding(e *models.Embedding) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.embeddings[e.ID.String()] = cloneEmbedding(e)
	return nil
}

func (s *MemoryStore) GetCollectionByName(name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	collectionID, ok := s.collections[name]
	if !ok {
		return "", fmt.Errorf("collection not found")
	}
	return collectionID, nil
}
