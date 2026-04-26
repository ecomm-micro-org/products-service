package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Embedding struct {
	ID           uuid.UUID
	CollectionID string          `gorm:"type:uuid"`
	Embedding    pgvector.Vector `json:"embedding" gorm:"type:vector(1024)"`
	Document     string          `json:"content"`
	Cmetadata    json.RawMessage `json:"metadata" gorm:"type:jsonb"`
}

func NewEmbedding() *Embedding {
	return &Embedding{}
}
