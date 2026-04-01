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

type Controller struct {
	e *token.JWTExtractor
}

func NewController(secretKey string) *Controller {
	return &Controller{
		e: token.NewJWTExtractor(secretKey),
	}
}

func extractAccessToken(c fiber.Ctx) (string, error) {
	accessToken := c.Get("Authorization")
	if accessToken == "" || !strings.HasPrefix(accessToken, "Bearer") {
		return "", fmt.Errorf("inlvaid token")
	}

	accessToken = strings.Split(accessToken, " ")[1]
	return accessToken, nil
}

func (con *Controller) GetProductByID(c fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON("invalid id")
	}

	s := services.New()
	if err := s.GetProductByID(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("the product does not exist")
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(s.ProductResponseDTO)
}

func (con *Controller) GetProductsByIDs(c fiber.Ctx) error {
	var productIDs *[]uint

	if err := c.Bind().Body(&productIDs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("invalid request body %v", err))
	}

	s := services.New()
	products, err := s.GetProductsByIDs(*productIDs)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

func (con *Controller) CalculateTotalPrice(c fiber.Ctx) error {
	var orderItems *[]dto.OrderItem

	if err := c.Bind().Body(&orderItems); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("invalid request body %v", err))
	}

	if orderItems == nil {
		return c.Status(fiber.StatusBadRequest).JSON("order items must contain a product id and a quantity")
	}

	s := services.New()
	products, err := s.CalculateTotalPrice(*orderItems)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

func (con *Controller) SearchProductsByKeyword(c fiber.Ctx) error {
	ctx := context.Background()

	keyword := c.Params("filter")
	if keyword == "" {
		return c.Status(fiber.StatusBadRequest).JSON("filter not sent")
	}

	s := services.New()

	products, err := s.SearchProductsByKeyword(ctx, keyword)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("filter did not match any product")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("internal server error")
	}

	return c.Status(fiber.StatusOK).JSON(products)
}

func (con *Controller) AddProduct(c fiber.Ctx) error {
	ctx := context.Background()
	s := services.New()

	if err := c.Bind().Body(&s.ProductRequestDTO); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("invlid json body %v", err))
	}

	accessToken, err := extractAccessToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(err)
	}

	uc, err := con.e.ExtractUserClaims(accessToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}
	s.UserClaims = uc

	if err := s.AddProduct(ctx); err != nil {
		fmt.Println("error inside add")
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(s.ProductResponseDTO)
}

func (con *Controller) UpdateProduct(c fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("invalid id %d", id))
	}

	s := services.New()
	if err := c.Bind().JSON(&s.ProductRequestDTO); err != nil {
		return err
	}

	accessToken := c.Get("Authorization")
	if accessToken == "" || strings.HasPrefix(accessToken, "Bearer ") {
		return c.Status(http.StatusUnauthorized).JSON("invalid token")
	}

	accessToken = strings.Split(accessToken, " ")[1]
	uc, err := con.e.ExtractUserClaims(accessToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}
	s.UserClaims = uc

	if err := s.UpdateProduct(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("product not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (con *Controller) DeleteProduct(c fiber.Ctx) error {
	ctx := context.Background()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fmt.Sprintf("invalid id %d", id))
	}

	s := services.New()
	accessToken := c.Get("Authorization")
	if accessToken == "" || strings.HasPrefix(accessToken, "Bearer ") {
		return c.Status(http.StatusUnauthorized).JSON("invalid token")
	}

	accessToken = strings.Split(accessToken, " ")[1]
	uc, err := con.e.ExtractUserClaims(accessToken)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err.Error())
	}
	s.UserClaims = uc

	if err := s.DeleteProduct(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON("product not found")
		}
		return c.Status(fiber.StatusInternalServerError).JSON("internal server error")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
