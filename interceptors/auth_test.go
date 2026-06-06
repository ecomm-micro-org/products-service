package interceptors

import (
	"context"
	"errors"
	"testing"

	internalauth "github.com/ecomm-micro-org/products-service/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeValidator struct {
	claims *internalauth.UserClaims
	err    error
}

func (f fakeValidator) ValidateToken(context.Context, string) (*internalauth.UserClaims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.claims, nil
}

func TestNewAuthInterceptorRejectsNilValidator(t *testing.T) {
	if _, err := NewAuthInterceptor(nil); err == nil {
		t.Fatal("expected error for nil validator")
	}
}

func TestUnaryAuthInterceptorAllowsPublicMethod(t *testing.T) {
	ai, err := NewAuthInterceptor(fakeValidator{})
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}

	called := false
	_, err = ai.UnaryAuthInterceptor()(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/products.ProductsService/GetProductByID"}, func(ctx context.Context, req any) (any, error) {
		called = true
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error = %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestUnaryAuthInterceptorRejectsMissingMetadata(t *testing.T) {
	ai, err := NewAuthInterceptor(fakeValidator{})
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}

	_, err = ai.UnaryAuthInterceptor()(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/products.ProductsService/AddProduct"}, func(context.Context, any) (any, error) {
		return nil, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestUnaryAuthInterceptorRejectsBadBearerPrefix(t *testing.T) {
	ai, err := NewAuthInterceptor(fakeValidator{})
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "Token abc"))
	_, err = ai.UnaryAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/products.ProductsService/AddProduct"}, func(context.Context, any) (any, error) {
		return nil, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestUnaryAuthInterceptorRejectsInvalidToken(t *testing.T) {
	ai, err := NewAuthInterceptor(fakeValidator{err: errors.New("bad token")})
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "Bearer abc"))
	_, err = ai.UnaryAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/products.ProductsService/AddProduct"}, func(context.Context, any) (any, error) {
		return nil, nil
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", status.Code(err))
	}
}

func TestUnaryAuthInterceptorAddsClaimsToContext(t *testing.T) {
	ai, err := NewAuthInterceptor(fakeValidator{claims: &internalauth.UserClaims{UserID: "user-1", Role: "seller"}})
	if err != nil {
		t.Fatalf("NewAuthInterceptor() error = %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("Authorization", "Bearer abc"))
	_, err = ai.UnaryAuthInterceptor()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/products.ProductsService/AddProduct"}, func(ctx context.Context, req any) (any, error) {
		if got := ctx.Value("userID"); got != "user-1" {
			t.Fatalf("expected userID in context, got %v", got)
		}
		if got := ctx.Value("role"); got != "seller" {
			t.Fatalf("expected role in context, got %v", got)
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error = %v", err)
	}
}
