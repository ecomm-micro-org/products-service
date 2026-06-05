package db

import (
	"github.com/ecomm-micro-org/products-service/models"
)

func AutoMigrate() {
	Client().AutoMigrate(&models.Product{})
}
