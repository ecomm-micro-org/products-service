package models

import (
	"strings"
	"testing"
)

func TestNewProduct(t *testing.T) {
	if got := NewProduct(); got == nil {
		t.Fatal("expected NewProduct to return a non-nil value")
	}
}

func TestProductString(t *testing.T) {
	p := &Product{
		Name:          "Keyboard",
		Price:         99.99,
		OriginalPrice: 119.99,
		Category:      "Electronics",
		Description:   "Mechanical keyboard",
		Rating:        5,
		Reviews:       10,
		Stock:         8,
		InStock:       true,
		Tags:          []string{"gaming", "mechanical"},
	}

	got := p.String()
	for _, want := range []string{"Keyboard", "Electronics", "Mechanical keyboard", "gaming"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q to appear in %q", want, got)
		}
	}
}
