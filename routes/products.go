package routes

import (
	"products/controllers"

	"github.com/gofiber/fiber/v3"
)

func ProductRoutes(r fiber.Router, c *controllers.Controller) {
	r.Get("/product/:id", c.GetProductByID)

	r.Post("/fetch_by_ids", c.GetProductsByIDs)
	r.Post("/add", c.AddProduct)
	r.Post("/calculate_total_price", c.CalculateTotalPrice)

	r.Put("/product/:id", c.UpdateProduct)

	r.Delete("/product/:id", c.DeleteProduct)
}
