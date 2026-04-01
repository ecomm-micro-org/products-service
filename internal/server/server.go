package server

import (
	"log"
	"os"
	"products/controllers"
	"products/internal/kafka"
	"products/internal/validate"
	"products/routes"
	"products/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

var app *fiber.App

func New() *fiber.App {
	return app
}

func consume() {
	errs := make(chan error, 10)

	s := services.New()

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

	go consume()

	defer app.Use(notFoundHandler)
	defer app.Use(recover.New())

	app.Get("/health", func(c fiber.Ctx) {
		c.SendStatus(fiber.StatusOK)
	})
	addRoutes(app)
}

func errorHandler(c fiber.Ctx, e error) error {
	err := e.Error()
	return c.Status(fiber.StatusInternalServerError).JSON(err)
}

var notFoundHandler = func(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotFound)
}

func addRoutes(app *fiber.App) {
	baseRouter := app.Group("/products")

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("no secret key defined")
	}

	c := controllers.NewController(secretKey)

	routes.ProductRoutes(baseRouter, c)
}
