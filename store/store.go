package store

import "github.com/ecomm-micro-org/products-service/models"

type Storer interface {
	AddProduct(p *models.Product) error
	SearchProductsByKeyword(keyword string) ([]*models.Product, error)
	GetProductByID(id uint64, p *models.Product) error
	GetProductsByIDs(productIDs []uint64) ([]*models.Product, error)
	UpdateProduct(p *models.Product) error
	DecreaseProductStock(id uint64, count uint64) (*models.Product, error)
	DeleteProduct(id uint64) error
	AddEmbedding(e *models.Embedding) error
	GetCollectionByName(name string) (string, error)
}
