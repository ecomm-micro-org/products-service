package interceptors

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type noopServerStream struct {
	ctx context.Context
}

func (s noopServerStream) SetHeader(metadata.MD) error  { return nil }
func (s noopServerStream) SendHeader(metadata.MD) error { return nil }
func (s noopServerStream) SetTrailer(metadata.MD)       {}
func (s noopServerStream) Context() context.Context     { return s.ctx }
func (s noopServerStream) SendMsg(any) error            { return nil }
func (s noopServerStream) RecvMsg(any) error            { return nil }

func TestUnaryLoggingInterceptor(t *testing.T) {
	li := NewLoggingInterceptor(zap.NewNop())

	called := false
	res, err := li.UnaryLoggingInterceptor()(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/products.ProductsService/GetProductByID"}, func(ctx context.Context, req any) (any, error) {
		called = true
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("UnaryLoggingInterceptor() error = %v", err)
	}
	if res != "ok" || !called {
		t.Fatalf("unexpected result: res=%v called=%v", res, called)
	}
}

func TestStreamLoggingInterceptor(t *testing.T) {
	li := NewLoggingInterceptor(zap.NewNop())

	called := false
	err := li.StreamLoggingInterceptor()(nil, noopServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{FullMethod: "/products.ProductsService/GetProductsByIDs", IsServerStream: true}, func(srv any, ss grpc.ServerStream) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("StreamLoggingInterceptor() error = %v", err)
	}
	if !called {
		t.Fatal("expected stream handler to be called")
	}
}
