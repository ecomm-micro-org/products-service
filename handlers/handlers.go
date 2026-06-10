package handlers

import (
	"context"
	"errors"
	"fmt"

	custom_errors "github.com/ecomm-micro-org/products-service/internal/constants/errors"
	"github.com/ecomm-micro-org/products-service/pb"
	"github.com/ecomm-micro-org/products-service/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type ProductHandler struct {
	pb.UnimplementedProductsServiceServer
	s *services.ProductService
}

func NewProductHandler(s *services.ProductService) *ProductHandler {
	return &ProductHandler{
		s: s,
	}
}

func (h *ProductHandler) GetProductByID(ctx context.Context, req *pb.GetProductByIDRequest) (*pb.GetProductByIDResponse, error) {
	res, err := h.s.GetProductByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

func (h *ProductHandler) GetProductsByIDs(ctx context.Context, req *pb.GetProductsByIDsRequest) (*pb.GetProductsByIDsResponse, error) {
	return h.s.GetProductsByIDs(ctx, req.ProductIds)
}

func (h *ProductHandler) AddProduct(ctx context.Context, req *pb.AddProductRequest) (*pb.AddProductResponse, error) {
	res, err := h.s.AddProduct(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("internal server error something went wrong : %v", err))
	}
	return res, nil
}

func (h *ProductHandler) CalculateTotalPrice(ctx context.Context, req *pb.CalculateTotalPriceRequest) (*pb.CalculateTotalPriceResponse, error) {
	res, err := h.s.CalculateTotalPrice(req.OrderItems)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return res, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*emptypb.Empty, error) {
	err := h.s.UpdateProduct(ctx, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		} else if errors.Is(err, custom_errors.ErrNotEnoughPermissions) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &emptypb.Empty{}, nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	err := h.s.DeleteProduct(ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		} else if errors.Is(err, custom_errors.ErrNotEnoughPermissions) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &emptypb.Empty{}, nil
}
