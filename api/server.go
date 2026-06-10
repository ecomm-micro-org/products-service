package api

import (
	"log"

	"github.com/ecomm-micro-org/products-service/handlers"
	"github.com/ecomm-micro-org/products-service/interceptors"
	"github.com/ecomm-micro-org/products-service/internal/auth"
	"github.com/ecomm-micro-org/products-service/internal/config"
	"github.com/ecomm-micro-org/products-service/pb"
	"github.com/ecomm-micro-org/products-service/services"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewServer(ps *services.ProductService) *grpc.Server {
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

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(li.UnaryLoggingInterceptor()),
		grpc.ChainUnaryInterceptor(ai.UnaryAuthInterceptor()),
		grpc.ChainStreamInterceptor(li.StreamLoggingInterceptor()),
	)

	pb.RegisterProductsServiceServer(grpcServer, handlers.NewProductHandler(ps))

	return grpcServer
}
