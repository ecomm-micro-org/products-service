package models

import (
	"errors"
	"fmt"
	"products/internal/database"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint `gorm:"primaryKey"`
	Name          string
	Price         float64
	OriginalPrice float64
	Image         string
	Category      string
	Description   string
	Rating        uint
	Reviews       uint
	Stock         uint64
	InStock       bool
	Tags          []string `gorm:"type:jsonb;serializer:json"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func New() *Product {
	return &Product{}
}

func (p *Product) String() string {
	return fmt.Sprintf("Name: %s\nPrice: %0.3f\nOriginalPrice: %0.3f\nCategory: %s\nDescription: %s,Rating: %d\nReviews: %d\nStock: %d\nInStock: %v\nTags: %v", p.Name, p.Price, p.OriginalPrice, p.Category, p.Description, p.Rating, p.Reviews, p.Stock, p.InStock, p.Tags)
}

func (p *Product) GetProductByID() error {
	if err := database.Client().First(p, p.ID); err != nil {
		return err.Error
	}
	return nil
}

func (p *Product) GetProductsByIDs(productIDs []uint) ([]Product, error) {
	var products []Product

	if err := database.Client().Where(productIDs).Find(&products).Order("id"); err.Error != nil {
		return nil, err.Error
	}

	return products, nil
}

func (p *Product) SearchProductsByKeyword(keyword string) ([]Product, error) {
	var products []Product

	if err := database.Client().Where("name ILIKE ? OR category ILIKE ? OR description ILIKE ?", keyword, keyword, keyword).Find(&products); err != nil {
		return nil, err.Error
	}

	return products, nil
}

func (p *Product) AddProduct() error {
	if err := database.Client().Save(&p); err != nil {
		return err.Error
	}
	return nil
}

func (p *Product) UpdateProduct() error {
	if err := database.Client().Save(&p); err != nil {
		return err.Error
	}
	return nil
}

func (p *Product) DecreaseProductStock(count uint) error {

	tx := database.Client().Begin()

	if err := tx.First(p, p.ID); err.Error != nil {
		tx.Rollback()
		return err.Error
	}

	if p.Stock < uint64(count) {
		tx.Rollback()
		return errors.New("unable to place order due to insufficient stock")
	}

	p.Stock = p.Stock - uint64(count)
	if p.Stock == 0 {
		p.InStock = false
	}

	result := tx.Save(p)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	tx.Commit()
	return nil
}

func (p *Product) DeleteProduct() error {
	if err := database.Client().Delete(&p); err != nil {
		return err.Error
	}
	return nil
}
