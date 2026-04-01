package migrations

import (
	"products/internal/database"
	"products/models"
)

func AutoMigrate() {
	database.Client().AutoMigrate(&models.Product{})
}
