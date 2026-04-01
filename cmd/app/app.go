package app

import (
	"log"
	"products/internal/cache"
	"products/internal/config"
	"products/internal/database"
	"products/internal/migrations"
	"products/internal/server"
)

func SetUp() {
	database.Connect()
	cache.Connect()
	migrations.AutoMigrate()

	server.SetUp()

	app := server.New()

	port := config.Config().Port
	if err := app.Listen(port); err != nil {
		log.Fatalf("err : %v", err)
	}
}
