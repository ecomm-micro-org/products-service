package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ecomm-micro-org/products-service/cache"
	"github.com/ecomm-micro-org/products-service/gen/pb"
	"github.com/ecomm-micro-org/products-service/models"
	"github.com/ecomm-micro-org/products-service/store"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
	"google.golang.org/grpc"
)

type ProductService struct {
	store       store.Storer
	genaiClient *genai.Client
	// messageService *MessageService
}

func NewProductService(
	store store.Storer,
	genaiClient *genai.Client,
	// , messageService *MessageService
) *ProductService {
	return &ProductService{
		store:       store,
		genaiClient: genaiClient,
		// messageService: messageService,
	}
}

func (p *ProductService) GetProductByID(ctx context.Context, id uint64) (*pb.GetProductByIDResponse, error) {
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

	res := &pb.GetProductByIDResponse{}

	res.Id = m.ID
	res.Name = m.Name
	res.Price = m.Price
	res.OriginalPrice = m.OriginalPrice
	res.Image = m.Image
	res.Category = m.Category
	res.Description = m.Description
	res.Stock = m.Stock
	res.InStock = m.InStock
	res.Rating = m.Rating
	res.Reviews = m.Reviews
	res.Tags = m.Tags

	return res, nil
}

func (p *ProductService) GetProductsByIDs(productIDs []uint64, stream grpc.ServerStreamingServer[pb.GetProductsByIDsResponse]) error {
	products, err := p.store.GetProductsByIDs(productIDs)
	if err != nil {
		return err
	}

	for _, v := range products {
		res := &pb.GetProductsByIDsResponse{}

		res.Id = v.ID
		res.Name = v.Name
		res.Price = v.Price
		res.OriginalPrice = v.OriginalPrice
		res.Image = v.Image
		res.Category = v.Category
		res.Description = v.Description
		res.Stock = v.Stock
		res.InStock = v.InStock
		res.Rating = v.Rating
		res.Reviews = v.Reviews
		res.Tags = v.Tags

		if err = stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}

func (p *ProductService) CalculateTotalPrice(orderItems []*pb.OrderItems) (*pb.CalculateTotalPriceResponse, error) {
	var productIDs []uint64
	for _, v := range orderItems {
		productIDs = append(productIDs, v.ProductId)
	}

	products, err := p.store.GetProductsByIDs(productIDs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[uint64]*models.Product)
	for _, product := range products {
		productMap[product.ID] = product
	}

	res := &pb.CalculateTotalPriceResponse{}
	for _, item := range orderItems {
		product, exists := productMap[item.ProductId]
		if !exists {
			return nil, fmt.Errorf("product with ID %d not found", item.ProductId)
		}

		// TODO: check if ts throws a SIGSEV,it mostly wont but check
		res.TotalPrice += product.Price * float64(item.Quantity)
	}

	return res, nil
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

func (p *ProductService) AddProduct(ctx context.Context, req *pb.AddProductRequest) (*pb.AddProductResponse, error) {
	if ctx.Value("role") == nil && ctx.Value("role") != "seller" {
		return nil, fmt.Errorf("cannot list product as user is not a seller")
	}

	m := models.NewProduct()

	m.Name = req.Name
	m.Price = req.Price
	m.Image = req.Image
	m.Category = req.Category
	m.Description = req.Description
	m.Stock = req.Stock
	m.InStock = req.InStock
	m.Tags = req.Tags

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

	res := &pb.AddProductResponse{}

	res.Id = m.ID
	res.Name = m.Name
	res.Price = m.Price
	res.OriginalPrice = m.OriginalPrice
	res.Image = m.Image
	res.Category = m.Category
	res.Description = m.Description
	res.Stock = m.Stock
	res.InStock = m.InStock
	res.Rating = m.Rating
	res.Reviews = m.Reviews
	res.Tags = m.Tags

	// TODO: run ts seperately in a goroutine
	// send req via chan
	embeddings, err := generateEmbedding(ctx, p.genaiClient, res)
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

	// go p.messageService.SendMessage(m.ID, m.Name, m.OriginalPrice)

	return res, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) error {
	if ctx.Value("role") == nil || ctx.Value("role") != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}

	m := models.NewProduct()

	m.ID = req.Id
	m.Name = req.Name
	m.Price = req.Price
	m.OriginalPrice = req.OriginalPrice
	m.Image = req.Image
	m.Category = req.Category
	m.Description = req.Description
	m.Stock = req.Stock
	m.InStock = req.InStock
	m.Tags = req.Tags

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

func (p *ProductService) DeleteProduct(ctx context.Context, id uint64) error {
	if ctx.Value("role") == nil || ctx.Value("role") != "seller" {
		return fmt.Errorf("cannot list product as user is not a seller")
	}

	cacheKey := fmt.Sprintf("product:%v", id)
	cache.Client().Del(ctx, cacheKey)
	return p.store.DeleteProduct(id)
}
