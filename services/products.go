package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"products/internal/cache"
	"products/internal/dto"
	"products/internal/token"
	"products/models"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
)

type Storer interface {
	AddProduct(p *models.Product) error
	SearchProductsByKeyword(keyword string) ([]*models.Product, error)
	GetProductByID(id uint, p *models.Product) error
	GetProductsByIDs(productIDs []uint) ([]*models.Product, error)
	UpdateProduct(p *models.Product) error
	DecreaseProductStock(id uint, count uint) (*models.Product, error)
	DeleteProduct(id uint) error
	AddEmbedding(e *models.Embedding) error
	GetCollectionByName(name string) (string, error)
}

type ProductService struct {
	UserClaims     *token.UserClaims
	store          Storer
	genaiClient    *genai.Client
	messageService *MessageService
}

func NewProductService(store Storer, genaiClient *genai.Client, messageService *MessageService) *ProductService {
	return &ProductService{
		store:          store,
		genaiClient:    genaiClient,
		messageService: messageService,
	}
}

func (p *ProductService) GetProductByID(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	m := models.NewProduct()

	cacheKey := fmt.Sprintf("product:%v", id)
	val, err := cache.Client().Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(val), &m); err != nil {
			return nil, err
		}
	} else {
		if err := p.store.GetProductByID(id, m); err != nil {
			return nil, err
		}

		cacheKey := fmt.Sprintf("product:%v", m.ID)
		cacheData, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}

		cache.Client().Set(ctx, cacheKey, cacheData, 10*time.Minute)
	}

	productResponse := &dto.ProductResponseDTO{}
	productResponse.ID = m.ID
	productResponse.Name = m.Name
	productResponse.Price = m.Price
	productResponse.OriginalPrice = m.OriginalPrice
	productResponse.Image = m.Image
	productResponse.Category = m.Category
	productResponse.Description = m.Description
	productResponse.Stock = m.Stock
	productResponse.InStock = m.InStock
	productResponse.Rating = m.Rating
	productResponse.Reviews = m.Reviews
	productResponse.Tags = m.Tags

	return productResponse, nil
}

func (p *ProductService) GetProductsByIDs(productIDs []uint) ([]dto.ProductResponseDTO, error) {
	products, err := p.store.GetProductsByIDs(productIDs)
	if err != nil {
		return nil, err
	}

	var productResponses []dto.ProductResponseDTO

	for _, v := range products {
		productResponse := &dto.ProductResponseDTO{}

		productResponse.ID = v.ID
		productResponse.Name = v.Name
		productResponse.Price = v.Price
		productResponse.OriginalPrice = v.OriginalPrice
		productResponse.Image = v.Image
		productResponse.Category = v.Category
		productResponse.Description = v.Description
		productResponse.Stock = v.Stock
		productResponse.InStock = v.InStock
		productResponse.Rating = v.Rating
		productResponse.Reviews = v.Reviews
		productResponse.Tags = v.Tags

		productResponses = append(productResponses, *productResponse)
	}

	return productResponses, nil
}

func (p *ProductService) CalculateTotalPrice(orderItems []dto.OrderItem) (float64, error) {
	var productIDs []uint
	for _, v := range orderItems {
		productIDs = append(productIDs, v.ProductID)
	}

	products, err := p.store.GetProductsByIDs(productIDs)
	if err != nil {
		return 0, err
	}

	productMap := make(map[uint]*models.Product)
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

func (p *ProductService) SearchProductsByKeyword(ctx context.Context, keyword string) ([]*dto.ProductResponseDTO, error) {
	var products []*models.Product

	val, err := cache.Client().Get(ctx, keyword).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(val), &products); err != nil {
			return nil, err
		}
	} else {
		products, err = p.store.SearchProductsByKeyword(keyword)
		if err != nil {
			return nil, err
		}

		cacheData, err := json.Marshal(products)
		if err != nil {
			return nil, err
		}

		if cmd := cache.Client().Set(ctx, keyword, cacheData, 30*time.Minute); cmd.Err() != nil {
			log.Println(cmd.Err())
		}
	}

	var productResponses []*dto.ProductResponseDTO

	for _, v := range products {
		productResponse := &dto.ProductResponseDTO{}

		productResponse.ID = v.ID
		productResponse.Name = v.Name
		productResponse.Price = v.Price
		productResponse.OriginalPrice = v.OriginalPrice
		productResponse.Image = v.Image
		productResponse.Category = v.Category
		productResponse.Description = v.Description
		productResponse.Stock = v.Stock
		productResponse.InStock = v.InStock
		productResponse.Rating = v.Rating
		productResponse.Reviews = v.Reviews
		productResponse.Tags = v.Tags

		productResponses = append(productResponses, productResponse)
	}

	return productResponses, nil
}

func structToText(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func generateEmbedding(ctx context.Context, client *genai.Client, data any) ([]float32, error) {
	text, err := structToText(data)
	if err != nil {
		return nil, err
	}

	embeddings, err := client.Models.EmbedContent(
		ctx,
		"gemini-embedding-2-preview",
		genai.Text(text),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return embeddings.Embeddings[0].Values, nil
}

func (p *ProductService) AddProduct(ctx context.Context, productRequest *dto.ProductRequestDTO) (*dto.ProductResponseDTO, error) {
	if p.UserClaims.Role != "seller" {
		return nil, fmt.Errorf("cannot list product as user is not a seller")
	}

	m := models.NewProduct()

	m.Name = productRequest.Name
	m.Price = productRequest.Price
	m.OriginalPrice = productRequest.OriginalPrice
	m.Image = productRequest.Image
	m.Category = productRequest.Category
	m.Description = productRequest.Description
	m.Stock = productRequest.Stock
	m.InStock = productRequest.InStock
	m.Tags = productRequest.Tags

	if err := p.store.AddProduct(m); err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("product:%v", m.ID)
	cacheData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if cmd := cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute); cmd.Err() != nil {
		log.Printf("unable to save product in cache : %v", err)
	}

	productResponse := &dto.ProductResponseDTO{}
	productResponse.ID = m.ID
	productResponse.Name = m.Name
	productResponse.Price = m.Price
	productResponse.OriginalPrice = m.OriginalPrice
	productResponse.Image = m.Image
	productResponse.Category = m.Category
	productResponse.Description = m.Description
	productResponse.Stock = m.Stock
	productResponse.InStock = m.InStock
	productResponse.Rating = m.Rating
	productResponse.Reviews = m.Reviews
	productResponse.Tags = m.Tags

	embeddings, err := generateEmbedding(ctx, p.genaiClient, productResponse)
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(map[string]any{
		"productId": m.ID,
	})
	if err != nil {
		return nil, err
	}

	collectionID, err := p.store.GetCollectionByName("products")
	if err != nil {
		return nil, err
	}

	e := models.NewEmbedding()
	e.ID = uuid.New()
	e.CollectionID = collectionID
	e.Embedding = pgvector.NewVector(embeddings)
	e.Document = m.String()
	e.Cmetadata = json.RawMessage(metadata)
	if err := p.store.AddEmbedding(e); err != nil {
		return nil, err
	}

	go p.messageService.SendMessage(m.ID, m.Name, m.OriginalPrice)

	return productResponse, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uint, productRequest *dto.ProductRequestDTO) error {
	if p.UserClaims.Role != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}

	m := models.NewProduct()

	m.ID = id
	m.Name = productRequest.Name
	m.Price = productRequest.Price
	m.OriginalPrice = productRequest.OriginalPrice
	m.Image = productRequest.Image
	m.Category = productRequest.Category
	m.Description = productRequest.Description
	m.Stock = productRequest.Stock
	m.InStock = productRequest.InStock
	m.Tags = productRequest.Tags

	if err := p.store.UpdateProduct(m); err != nil {
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
	m, err := p.store.DecreaseProductStock(id, count)
	if err != nil {
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

// func (p *ProductService) IncreaseProductStock(ctx context.Context, id uint, count uint) error {
// 	if err := m.DecreaseProductStock(count); err != nil {
// 		return err
// 	}
//
// 	cacheKey := fmt.Sprintf("product:%v", m.ID)
// 	cacheData, err := json.Marshal(m)
// 	if err != nil {
// 		return err
// 	}
//
// 	cache.Client().Set(ctx, cacheKey, cacheData, 30*time.Minute)
//
// 	return nil
// }

func (p *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	if p.UserClaims.Role != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}

	cacheKey := fmt.Sprintf("product:%v", id)
	cache.Client().Del(ctx, cacheKey)
	return p.store.DeleteProduct(id)
}
