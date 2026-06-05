package api

import (
	"context"
	"log"

	"github.com/ecomm-micro-org/products-service/auth"
	"github.com/ecomm-micro-org/products-service/db"
	"github.com/ecomm-micro-org/products-service/gen/pb"
	"github.com/ecomm-micro-org/products-service/handlers"
	"github.com/ecomm-micro-org/products-service/interceptors"
	"github.com/ecomm-micro-org/products-service/internal/config"
	"github.com/ecomm-micro-org/products-service/services"
	"github.com/ecomm-micro-org/products-service/store"
	"go.uber.org/zap"
	"google.golang.org/genai"
	"google.golang.org/grpc"
)

func NewServer() *grpc.Server {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("unable to create zap logger instance\n")
	}

	am, err := auth.NewAuthManager(config.Config().AuthSecretKey)
	if err != nil {
		log.Fatalf("unable to create auth manager")
	}

	// interceptors
	li := interceptors.NewLoggingInterceptor(l)
	ai, err := interceptors.NewAuthInterceptor(am)
	if err != nil {
		log.Fatalf("unable to create auth interceptor")
	}

	s := store.NewPGStore(db.Client(), config.Config().EmbeddingTableName, config.Config().EmbeddingCollectionTableName)

	gc, err := genai.NewClient(
		context.Background(),
		&genai.ClientConfig{
			APIKey:  config.Config().GeminiAPIKey,
			Backend: genai.BackendGeminiAPI,
		},
	)
	if err != nil {
		log.Fatalf("unable to create genai client instance : %v\n", err)
	}

	ps := services.NewProductService(s, gc)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(li.UnaryLoggingInterceptor()),
		grpc.ChainUnaryInterceptor(ai.UnaryAuthInterceptor()),
		grpc.ChainStreamInterceptor(li.StreamLoggingInterceptor()),
	)

	pb.RegisterProductsServiceServer(grpcServer, handlers.NewProductHandler(ps))

	return grpcServer
}
