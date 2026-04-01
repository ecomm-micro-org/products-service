package models

import (
	"context"
	"encoding/json"
	"products/internal/database"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Embedding struct {
	ID        uuid.UUID       `gorm:"type:uuid"`
	Embedding pgvector.Vector `json:"embedding" gorm:"type:vector(1024)"`
	Content   string          `json:"content"`
	Metadata  json.RawMessage `json:"metadata" gorm:"type:jsonb"`
}

func NewEmbedding() *Embedding {
	return &Embedding{ID: uuid.New()}
}

func (e *Embedding) AddEmbedding(ctx context.Context) error {
	return database.Client().Table("vector_store").Save(&e).Error
}
