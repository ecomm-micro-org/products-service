package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint64 `gorm:"primaryKey"`
	Name          string
	Price         float64
	OriginalPrice float64
	Image         string
	Category      string
	Description   string
	Rating        uint64
	Reviews       uint64
	Stock         uint64
	InStock       bool
	Tags          []string `gorm:"type:jsonb;serializer:json"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func NewProduct() *Product {
	return &Product{}
}

func (p *Product) String() string {
	return fmt.Sprintf("Name: %s\nPrice: %0.3f\nOriginalPrice: %0.3f\nCategory: %s\nDescription: %s,Rating: %d\nReviews: %d\nStock: %d\nInStock: %v\nTags: %v", p.Name, p.Price, p.OriginalPrice, p.Category, p.Description, p.Rating, p.Reviews, p.Stock, p.InStock, p.Tags)
}
