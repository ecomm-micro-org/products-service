package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/ecomm-micro-org/products-service/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	Validator interface {
		ValidateToken(context.Context, string) (*auth.UserClaims, error)
	}

	authInterceptor struct {
		validator Validator
	}
)

func NewAuthInterceptor(validator Validator) (*authInterceptor, error) {
	if validator == nil {
		return nil, fmt.Errorf("validator cannot be nil")
	}

	return &authInterceptor{
		validator: validator,
	}, nil
}

func (ai *authInterceptor) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	publicUrls := map[string]bool{
		"/products.ProductsService/GetProductByID":   true,
		"/products.ProductsService/GetProductsByIDs": true,
	}

	return func(ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if publicUrls[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not provided")
		}

		token := md.Get("Authorization")
		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, "no 'Authorization' token provided")
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(token[0], prefix) {
			return nil, status.Error(codes.Unauthenticated, "'Authorization' header must start with 'Bearer '")
		}
		tokenStr := strings.TrimPrefix(token[0], prefix)

		claims, err := ai.validator.ValidateToken(ctx, tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		ctx = contextWithClaims(ctx, claims)
		return handler(ctx, req)
	}
}

func contextWithClaims(ctx context.Context, claims *auth.UserClaims) context.Context {
	ctx = context.WithValue(ctx, "userID", claims.UserID)
	ctx = context.WithValue(ctx, "role", claims.Role)
	return ctx
}
