package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"products/internal/dto"
	"products/internal/token"
	"products/services"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type ErrorMsg struct {
	message string `json:"message"`
}

type Controller struct {
	e *token.JWTExtractor
	s *services.ProductService
}

func NewController(secretKey string, s *services.ProductService) *Controller {
	return &Controller{
		e: token.NewJWTExtractor(secretKey),
		s: s,
	}
}

func extractAccessToken(c fiber.Ctx) (string, error) {
	accessToken := c.Get("Authorization")
	if accessToken == "" || !strings.HasPrefix(accessToken, "Bearer") {
		return "", fmt.Errorf("invalid token")
	}

	accessToken = strings.Split(accessToken, " ")[1]
	return accessToken, nil
}

// GetProductByID Fetches a product by its ID
//
//	@Summary      Fetch Product
//	@Description  Fetches an product by its id
//	@Tags         product
//	@Produce      json
//	@Param        id    path     int  true "product search by id"
//	@Success      200  {object}  dto.ProductResponseDTO
//	@Failure      400  {object}  ErrorMsg
//	@Failure      404  {object}  ErrorMsg
//	@Failure      500  {string}  string
//	@Router       /product/{id} [get]
func (con *Controller) GetProductByID(c fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: "invalid id"})
	}

	p, err := con.s.GetProductByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorMsg{message: "the product does not exist"})
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(p)
}

// GetProductsByIDs Fetches products by their IDs
//
//	@Summary      Fetch Products by IDs
//	@Description  Fetches products by an array of ids
//	@Tags         product
//	@Accept       json
//	@Produce      json
//	@Param        productIDs    body     []uint  true "array of product ids"
//	@Success      200  {array}   dto.ProductResponseDTO
//	@Failure      400  {object}  ErrorMsg
//	@Failure      500  {string}  string
//	@Router       /fetch_by_ids [post]
func (con *Controller) GetProductsByIDs(c fiber.Ctx) error {
	var productIDs *[]uint

	if err := c.Bind().Body(&productIDs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: fmt.Sprintf("invalid request body %v", err)})
	}

	products, err := con.s.GetProductsByIDs(*productIDs)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

// CalculateTotalPrice Calculates the total price of an order
//
//	@Summary      Calculate Total Price
//	@Description  Calculates the total price of order items
//	@Tags         product
//	@Accept       json
//	@Produce      json
//	@Param        orderItems    body     []dto.OrderItem  true "array of order items"
//	@Success      200  {array}   dto.ProductResponseDTO
//	@Failure      400  {object}  ErrorMsg
//	@Failure      500  {object}  ErrorMsg
//	@Router       /calculate_total_price [post]
func (con *Controller) CalculateTotalPrice(c fiber.Ctx) error {
	var orderItems *[]dto.OrderItem

	if err := c.Bind().Body(&orderItems); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: fmt.Sprintf("invalid request body %v", err)})
	}

	if orderItems == nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: "order items must contain a product id and a quantity"})
	}

	products, err := con.s.CalculateTotalPrice(*orderItems)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorMsg{message: err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

// SearchProductsByKeyword Searches products by keyword
//
//	@Summary      Search Products
//	@Description  Searches products by keyword
//	@Tags         product
//	@Produce      json
//	@Param        filter    path     string  true "keyword to search"
//	@Success      200  {array}   dto.ProductResponseDTO
//	@Failure      400  {object}  ErrorMsg
//	@Failure      404  {object}  ErrorMsg
//	@Failure      500  {object}  ErrorMsg
//	@Router       /search/{filter} [get]
func (con *Controller) SearchProductsByKeyword(c fiber.Ctx) error {
	ctx := context.Background()

	keyword := c.Params("filter")
	if keyword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: "filter not sent"})
	}

	products, err := con.s.SearchProductsByKeyword(ctx, keyword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorMsg{message: "filter did not match any product"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorMsg{message: "internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

// AddProduct Adds a new product
//
//	@Summary      Add Product
//	@Description  Adds a new product
//	@Tags         product
//	@Accept       json
//	@Produce      json
//	@Param        product    body     dto.ProductRequestDTO  true "product details"
//	@Success      201  {object}  dto.ProductResponseDTO
//	@Failure      400  {object}  ErrorMsg
//	@Failure      401  {object}  ErrorMsg
//	@Failure      500  {object}  ErrorMsg
//	@Security     BearerAuth
//	@Router       /add [post]
func (con *Controller) AddProduct(c fiber.Ctx) error {
	ctx := context.Background()

	productRequest := &dto.ProductRequestDTO{}

	if err := c.Bind().Body(productRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: fmt.Sprintf("invlid json body %v", err)})
	}

	accessToken, err := extractAccessToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorMsg{message: err.Error()})
	}

	uc, err := con.e.ExtractUserClaims(accessToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(ErrorMsg{message: err.Error()})
	}
	con.s.UserClaims = uc

	p, err := con.s.AddProduct(ctx, productRequest)
	if err != nil {
		fmt.Printf("%v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorMsg{message: err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(p)
}

// UpdateProduct Updates an existing product
//
//	@Summary      Update Product
//	@Description  Updates an existing product by id
//	@Tags         product
//	@Accept       json
//	@Produce      json
//	@Param        id    path     int  true "product id"
//	@Param        product    body     dto.ProductRequestDTO  true "product details"
//	@Success      204  {string}  string "No Content"
//	@Failure      400  {object}  ErrorMsg
//	@Failure      401  {object}  ErrorMsg
//	@Failure      404  {object}  ErrorMsg
//	@Failure      500  {object}  ErrorMsg
//	@Security     BearerAuth
//	@Router       /product/{id} [put]
func (con *Controller) UpdateProduct(c fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: fmt.Sprintf("invalid id %d", id)})
	}

	productRequest := &dto.ProductRequestDTO{}
	if err := c.Bind().JSON(productRequest); err != nil {
		return err
	}

	accessToken, err := extractAccessToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorMsg{message: err.Error()})
	}
	uc, err := con.e.ExtractUserClaims(accessToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(ErrorMsg{message: err.Error()})
	}
	con.s.UserClaims = uc

	if err := con.s.UpdateProduct(ctx, uint(id), productRequest); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorMsg{message: "product not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorMsg{message: "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteProduct Deletes a product by ID
//
//	@Summary      Delete Product
//	@Description  Deletes a product by id
//	@Tags         product
//	@Param        id    path     int  true "product id"
//	@Success      204  {string}  string "No Content"
//	@Failure      400  {object}  ErrorMsg
//	@Failure      401  {object}  ErrorMsg
//	@Failure      404  {object}  ErrorMsg
//	@Failure      500  {object}  ErrorMsg
//	@Security     BearerAuth
//	@Router       /product/{id} [delete]
func (con *Controller) DeleteProduct(c fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorMsg{message: fmt.Sprintf("invalid id %d", id)})
	}

	accessToken, err := extractAccessToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorMsg{message: err.Error()})
	}

	uc, err := con.e.ExtractUserClaims(accessToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(ErrorMsg{message: err.Error()})
	}
	con.s.UserClaims = uc

	if err := con.s.DeleteProduct(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorMsg{message: "product not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorMsg{message: "internal server error"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
