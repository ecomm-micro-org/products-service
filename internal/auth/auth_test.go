package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewAuthManagerRejectsEmptySecret(t *testing.T) {
	if _, err := NewAuthManager(""); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestValidateTokenSuccess(t *testing.T) {
	am, err := NewAuthManager("secret")
	if err != nil {
		t.Fatalf("NewAuthManager() error = %v", err)
	}

	claims := NewUserClaims("user-1", "user@example.com", "user", "seller", time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	got, err := am.ValidateToken(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if got.UserID != claims.UserID || got.Role != "seller" {
		t.Fatalf("unexpected claims: %+v", got)
	}
}

func TestValidateTokenRejectsInvalidToken(t *testing.T) {
	am, err := NewAuthManager("secret")
	if err != nil {
		t.Fatalf("NewAuthManager() error = %v", err)
	}

	if _, err := am.ValidateToken(context.Background(), "not-a-jwt"); err == nil {
		t.Fatal("expected invalid token error")
	}
}
