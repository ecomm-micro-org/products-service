package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"products/internal/cache"
	"products/internal/config"
	"products/internal/dto"
	"products/internal/token"
	"products/models"
	"time"

	"github.com/pgvector/pgvector-go"
)

type ProductService struct {
	UserClaims         *token.UserClaims
	ProductRequestDTO  dto.ProductRequestDTO
	ProductResponseDTO dto.ProductResponseDTO
}

func New() *ProductService {
	return &ProductService{}
}

func (p *ProductService) generateTextEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddingModelURL := config.Config().EmbeddingModelURL

	reqData := struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}{
		Model:  config.Config().EmbeddingModelName,
		Prompt: text,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(embeddingModelURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var resData struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(body, &resData); err != nil {
		return nil, err
	}

	return resData.Embedding, nil
}

func (p *ProductService) GetProductByID(ctx context.Context, id uint) error {
	m := models.New()

	cacheKey := fmt.Sprintf("product:%v", id)
	val, err := cache.Client().Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(val), &m); err != nil {
			return err
		}
	} else {
		m.ID = id
		if err := m.GetProductByID(); err != nil {
			return err
		}

		cacheKey := fmt.Sprintf("product:%v", m.ID)
		cacheData, err := json.Marshal(m)
		if err != nil {
			return err
		}

		cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute)
	}

	p.ProductResponseDTO.ID = m.ID
	p.ProductResponseDTO.Name = m.Name
	p.ProductResponseDTO.Price = m.Price
	p.ProductResponseDTO.OriginalPrice = m.OriginalPrice
	p.ProductResponseDTO.Image = m.Image
	p.ProductResponseDTO.Category = m.Category
	p.ProductResponseDTO.Description = m.Description
	p.ProductResponseDTO.Stock = m.Stock
	p.ProductResponseDTO.InStock = m.InStock
	p.ProductResponseDTO.Rating = m.Rating
	p.ProductResponseDTO.Reviews = m.Reviews
	p.ProductResponseDTO.Tags = m.Tags

	return nil
}

func (p *ProductService) GetProductsByIDs(productIDs []uint) ([]dto.ProductResponseDTO, error) {
	m := models.New()

	products, err := m.GetProductsByIDs(productIDs)
	if err != nil {
		return nil, err
	}

	var productResponses []dto.ProductResponseDTO

	for _, v := range products {
		productResponseDTO := &dto.ProductResponseDTO{}

		productResponseDTO.ID = v.ID
		productResponseDTO.Name = v.Name
		productResponseDTO.Price = v.Price
		productResponseDTO.OriginalPrice = v.OriginalPrice
		productResponseDTO.Image = v.Image
		productResponseDTO.Category = v.Category
		productResponseDTO.Description = v.Description
		productResponseDTO.Stock = v.Stock
		productResponseDTO.InStock = v.InStock
		productResponseDTO.Rating = v.Rating
		productResponseDTO.Reviews = v.Reviews
		productResponseDTO.Tags = v.Tags

		productResponses = append(productResponses, *productResponseDTO)
	}

	return productResponses, nil
}

func (p *ProductService) CalculateTotalPrice(orderItems []dto.OrderItem) (float64, error) {
	m := models.New()

	var productIDs []uint
	for _, v := range orderItems {
		productIDs = append(productIDs, v.ProductID)
	}

	products, err := m.GetProductsByIDs(productIDs)
	if err != nil {
		return 0, err
	}

	productMap := make(map[uint]models.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	var total float64
	for _, item := range orderItems {
		product, exists := productMap[item.ProductID]
		if !exists {
			return 0, fmt.Errorf("product with ID %d not found", item.ProductID)
		}
		total += product.Price * float64(item.Quantity)
	}

	return total, nil
}

func (p *ProductService) SearchProductsByKeyword(ctx context.Context, keyword string) ([]dto.ProductResponseDTO, error) {
	m := models.New()
	var products []models.Product

	val, err := cache.Client().Get(ctx, keyword).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(val), &products); err != nil {
			return nil, err
		}
	} else {
		products, err = m.SearchProductsByKeyword(keyword)
		if err != nil {
			return nil, err
		}

		cacheData, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}

		if cmd := cache.Client().Set(ctx, keyword, cacheData, 30*time.Minute); cmd.Err() != nil {
			return nil, cmd.Err()
		}
	}

	var productResponses []dto.ProductResponseDTO

	for _, v := range products {
		productResponseDTO := &dto.ProductResponseDTO{}

		productResponseDTO.ID = v.ID
		productResponseDTO.Name = v.Name
		productResponseDTO.Price = v.Price
		productResponseDTO.OriginalPrice = v.OriginalPrice
		productResponseDTO.Image = v.Image
		productResponseDTO.Category = v.Category
		productResponseDTO.Description = v.Description
		productResponseDTO.Stock = v.Stock
		productResponseDTO.InStock = v.InStock
		productResponseDTO.Rating = v.Rating
		productResponseDTO.Reviews = v.Reviews
		productResponseDTO.Tags = v.Tags

		productResponses = append(productResponses, *productResponseDTO)
	}

	return productResponses, nil
}

func (p *ProductService) AddProduct(ctx context.Context) error {
	if p.UserClaims.Role != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}

	m := models.New()

	m.Name = p.ProductRequestDTO.Name
	m.Price = p.ProductRequestDTO.Price
	m.OriginalPrice = p.ProductRequestDTO.OriginalPrice
	m.Image = p.ProductRequestDTO.Image
	m.Category = p.ProductRequestDTO.Category
	m.Description = p.ProductRequestDTO.Description
	m.Stock = p.ProductRequestDTO.Stock
	m.InStock = p.ProductRequestDTO.InStock
	m.Tags = p.ProductRequestDTO.Tags

	if err := m.AddProduct(); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("product:%v", m.ID)
	cacheData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if cmd := cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute); cmd.Err() != nil {
		log.Printf("unable to save product in cache : %v", err)
	}

	p.ProductResponseDTO.ID = m.ID
	p.ProductResponseDTO.Name = m.Name
	p.ProductResponseDTO.Price = m.Price
	p.ProductResponseDTO.OriginalPrice = m.OriginalPrice
	p.ProductResponseDTO.Image = m.Image
	p.ProductResponseDTO.Category = m.Category
	p.ProductResponseDTO.Description = m.Description
	p.ProductResponseDTO.Stock = m.Stock
	p.ProductResponseDTO.InStock = m.InStock
	p.ProductResponseDTO.Rating = m.Rating
	p.ProductResponseDTO.Reviews = m.Reviews
	p.ProductResponseDTO.Tags = m.Tags

	embedding, err := p.generateTextEmbedding(ctx, fmt.Sprintf("%v", m.String()))
	if err != nil {
		return err
	}

	metadata, err := json.Marshal(map[string]any{
		"productId": m.ID,
	})
	if err != nil {
		return err
	}

	e := models.NewEmbedding()
	e.Embedding = pgvector.NewVector(embedding)
	e.Content = m.String()
	e.Metadata = json.RawMessage(metadata)
	if err := e.AddEmbedding(ctx); err != nil {
		return err
	}

	return nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uint) error {
	if p.UserClaims.Role != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}

	m := models.New()

	m.ID = id
	m.Name = p.ProductRequestDTO.Name
	m.Price = p.ProductRequestDTO.Price
	m.OriginalPrice = p.ProductRequestDTO.OriginalPrice
	m.Image = p.ProductRequestDTO.Image
	m.Category = p.ProductRequestDTO.Category
	m.Description = p.ProductRequestDTO.Description
	m.Stock = p.ProductRequestDTO.Stock
	m.InStock = p.ProductRequestDTO.InStock
	m.Tags = p.ProductRequestDTO.Tags

	if err := m.UpdateProduct(); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("product:%v", m.ID)
	cacheData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if cmd := cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute); cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (p *ProductService) DecreaseProductStock(ctx context.Context, id uint, count uint) error {
	m := models.New()
	m.ID = id

	if err := m.DecreaseProductStock(count); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("product:%v", m.ID)
	cacheData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute)

	return nil
}

func (p *ProductService) IncreaseProductStock(ctx context.Context, id uint, count uint) error {
	m := models.New()
	m.ID = id

	if err := m.DecreaseProductStock(count); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("product:%v", m.ID)
	cacheData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute)

	return nil
}

func (p *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if p.UserClaims.Role != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}
	m := models.New()
	m.ID = id

	cacheKey := fmt.Sprintf("product:%v", id)
	cache.Client().Del(ctx, cacheKey)
	return m.DeleteProduct()
}
