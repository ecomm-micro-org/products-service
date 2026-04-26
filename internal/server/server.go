package server

import (
	"context"
	"log"
	"os"
	"products/controllers"
	"products/internal/database"
	"products/internal/kafka"
	"products/internal/validate"
	"products/routes"
	"products/services"
	"products/store"

	appConfig "products/internal/config"

	"github.com/go-playground/validator/v10"
	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/trycourier/courier-go/v4"
	"github.com/trycourier/courier-go/v4/option"
	"google.golang.org/genai"
)

var app *fiber.App

func New() *fiber.App {
	return app
}

func consume(s *services.ProductService) {
	errs := make(chan error, 10)

	kc := kafka.NewConsumer(kafka.TopicOrderCreated)
	go kc.Consume(s, errs)

	for err := range errs {
		log.Printf("error occurred while consuming on kafka : %v\n", err.Error())
	}

	kc.Close()
}

func SetUp() {
	config := fiber.Config{
		BodyLimit:    10 * 1024 * 1024,
		ErrorHandler: errorHandler,
		StructValidator: &validate.StructValidator{
			Validator: validator.New(),
		},
	}

	app = fiber.New(config)
	app.Use(logger.New())

	geminiAPIKey := appConfig.Config().GeminiAPIKey
	embeddingTableName := appConfig.Config().EmbeddingTableName
	embeddingCollectionTableName := appConfig.Config().EmbeddingCollectionTableName

	courierClient := courier.NewClient(option.WithAPIKey("COURIER_KEY"))

	store := store.NewPGStore(database.Client(), embeddingTableName, embeddingCollectionTableName)

	genaiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  geminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal("unable to create genai cleint")
	}
	ms := services.NewMessageService(&courierClient)

	s := services.NewProductService(
		store,
		genaiClient,
		ms,
	)

	go consume(s)

	defer app.Use(notFoundHandler)
	defer app.Use(recover.New())

	app.Get("/health", func(c fiber.Ctx) {
		c.SendStatus(fiber.StatusOK)
	})

	app.Get("/swagger/*", swaggo.HandlerDefault)

	baseRouter := app.Group("/products")

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("no secret key defined")
	}

	c := controllers.NewController(secretKey, s)

	routes.ProductRoutes(baseRouter, c)
}

func errorHandler(c fiber.Ctx, e error) error {
	err := e.Error()
	return c.Status(fiber.StatusInternalServerError).JSON(err)
}

var notFoundHandler = func(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotFound)
}
