package models

import "testing"

func TestNewEmbedding(t *testing.T) {
	if got := NewEmbedding(); got == nil {
		t.Fatal("expected NewEmbedding to return a non-nil value")
	}
}
