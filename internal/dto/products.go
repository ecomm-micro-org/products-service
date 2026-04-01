package dto

type ProductRequestDTO struct {
	Name          string `validate:"required"`
	Price         float64
	OriginalPrice float64 `json:"original_price" validate:"required"`
	Image         string  `json:"image" validate:"required"`
	Category      string  `json:"category" validate:"required"`
	Description   string  `json:"description" validate:"lte=250,gte=0"`
	Stock         uint64  `json:"stock" validate:"required"`
	InStock       bool    `json:"in_stock" validate:"required"`
	Tags          []string
}

type ProductResponseDTO struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Price         float64  `json:"price"`
	OriginalPrice float64  `json:"original_price"`
	Image         string   `json:"image"`
	Category      string   `json:"category"`
	Description   string   `json:"description"`
	Rating        uint     `json:"rating"`
	Reviews       uint     `json:"reviews"`
	Stock         uint64   `json:"stock"`
	InStock       bool     `json:"in_stock"`
	Tags          []string `json:"tags"`
}

type OrderItem struct {
	ProductID uint `json:"product_id"`
	Quantity  uint `json:"quantity"`
}
