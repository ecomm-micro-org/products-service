package store

import (
	"errors"
	"fmt"

	"github.com/ecomm-micro-org/products-service/models"
	"gorm.io/gorm"
)

type PGStore struct {
	db                           *gorm.DB
	embeddingTableName           string
	embeddingCollectionTableName string
}

func NewPGStore(db *gorm.DB, embeddingTableName, embeddingCollectionTableName string) *PGStore {
	return &PGStore{
		db:                           db,
		embeddingTableName:           embeddingTableName,
		embeddingCollectionTableName: embeddingCollectionTableName,
	}
}

func (s *PGStore) GetProductByID(id uint64, p *models.Product) error {
	if err := s.db.First(p, id); err != nil {
		return err.Error
	}
	return nil
}

func (s *PGStore) GetProductsByIDs(productIDs []uint64) ([]*models.Product, error) {
	var products []*models.Product

	if err := s.db.Where(productIDs).Find(&products).Order("id"); err.Error != nil {
		return nil, err.Error
	}

	return products, nil
}

func (s *PGStore) AddProduct(p *models.Product) error {
	if err := s.db.Save(&p); err != nil {
		return err.Error
	}
	return nil
}

func (s *PGStore) UpdateProduct(p *models.Product) error {
	if err := s.db.Save(&p); err != nil {
		return err.Error
	}
	return nil
}

func (s *PGStore) DecreaseProductStock(id uint64, count uint64) (*models.Product, error) {
	tx := s.db.Begin()
	p := models.NewProduct()

	if err := tx.First(p, id); err.Error != nil {
		tx.Rollback()
		return nil, err.Error
	}

	if p.Stock < count {
		tx.Rollback()
		return nil, errors.New("unable to place order due to insufficient stock")
	}

	p.Stock = p.Stock - count
	if p.Stock == 0 {
		p.InStock = false
	}

	result := tx.Save(p)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	tx.Commit()
	return p, nil
}

func (s *PGStore) DeleteProduct(id uint64) error {
	p := models.NewProduct()
	p.ID = id

	if err := s.db.Delete(&p); err != nil {
		return err.Error
	}
	return nil
}

func (s *PGStore) AddEmbedding(e *models.Embedding) error {
	return s.db.Table(s.embeddingTableName).Save(&e).Error
}

func (s *PGStore) GetCollectionByName(name string) (string, error) {
	var collectionID string
	if err := s.db.Table(s.embeddingCollectionTableName).Select("uuid").Where("name = ?", name).Find(&collectionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("collection not found")
		}
		return "", err
	}
	return collectionID, nil
}
