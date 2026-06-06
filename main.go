package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecomm-micro-org/products-service/api"
	"github.com/ecomm-micro-org/products-service/cache"
	"github.com/ecomm-micro-org/products-service/db"
	"github.com/ecomm-micro-org/products-service/internal/config"
	kafkaConsumer "github.com/ecomm-micro-org/products-service/internal/kafka"
	"github.com/ecomm-micro-org/products-service/services"
	"github.com/ecomm-micro-org/products-service/store"
	"github.com/segmentio/kafka-go"
	"google.golang.org/genai"
	"google.golang.org/grpc"
)

func main() {
	config.Init()

	db.Connect()
	db.AutoMigrate()
	cache.Connect()

	kafkaErr := make(chan error)
	cCfg := kafkaConsumer.ConsumerConfig{
		Brokers:        config.Config().Brokers,
		Topic:          kafkaConsumer.TopicOrderPlaced.String(),
		GroupID:        "products-service",
		Partition:      0,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 1000, // 1 second
		StartOffset:    kafka.FirstOffset,
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

	ps1 := services.NewProductService(s, gc)
	ps2 := services.NewProductService(s, gc)

	c := kafkaConsumer.NewConsumer(cCfg, ps1, kafkaErr)
	c.Init()
	go c.Consume()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	grpcServer := api.NewServer(ps2)

	if err := runGRPCServer(context.Background(), grpcServer, 3*time.Second); err != nil {
		log.Fatalf("unable to start the server : %v", err)
	}
}

func runGRPCServer(ctx context.Context, grpcServer *grpc.Server, shutdownTimeout time.Duration) error {
	serverErr := make(chan error, 1)

	go func() {
		log.Println("products service running on port", config.Config().Port)
		lis, err := net.Listen("tcp", config.Config().Port)
		if err != nil {
			log.Fatalf("unable to listen on port %s\n", config.Config().Port)
		}

		if err := grpcServer.Serve(lis); !errors.Is(err, grpc.ErrServerStopped) {
			serverErr <- err
		}
		close(serverErr)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-shutdown:
		log.Println("shutdown signal received")
	case <-ctx.Done():
		log.Println("parent context cancelled")
	}

	grpcServer.GracefulStop()

	log.Println("sever exited successfully")
	return nil
}
