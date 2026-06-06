package auth

import (
	"testing"
	"time"
)

func TestNewUserClaims(t *testing.T) {
	claims := NewUserClaims("user-1", "user@example.com", "risbern", "seller", time.Hour)

	if claims.UserID != "user-1" || claims.Email != "user@example.com" || claims.Username != "risbern" || claims.Role != "seller" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.RegisteredClaims.ID == "" {
		t.Fatal("expected token ID to be set")
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		t.Fatal("expected issued at and expiry to be set")
	}
	if !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		t.Fatal("expected expiry to be after issued-at")
	}
}
