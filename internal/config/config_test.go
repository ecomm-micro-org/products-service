package config

import (
	"sync"
	"testing"
)

func resetConfigForTest() {
	defaultConfig = nil
	once = sync.Once{}
}

func TestInitLoadsEnvironmentVariables(t *testing.T) {
	resetConfigForTest()
	t.Setenv("DSN", "postgres://dsn")
	t.Setenv("BROKERS", "localhost:9092,localhost:9093")
	t.Setenv("CACHE_ADDR", "localhost:6379")
	t.Setenv("CACHE_PASSWD", "secret")
	t.Setenv("PORT", ":42069")
	t.Setenv("EMBEDDING_COLLECTION_TABLE_NAME", "collections")
	t.Setenv("EMBEDDING_TABLE_NAME", "embeddings")
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	t.Setenv("SECRET_KEY", "jwt-secret")
	t.Setenv("LOG_FILE", "app.log")

	Init()
	cfg := Config()
	if cfg == nil {
		t.Fatal("expected config to be initialized")
	}
	if cfg.DSN != "postgres://dsn" || cfg.CacheAddr != "localhost:6379" || cfg.Port != ":42069" {
		t.Fatalf("unexpected config values: %+v", cfg)
	}
	if len(cfg.Brokers) != 2 || cfg.Brokers[0] != "localhost:9092" || cfg.Brokers[1] != "localhost:9093" {
		t.Fatalf("unexpected brokers: %v", cfg.Brokers)
	}
	if cfg.EmbeddingCollectionTableName != "collections" || cfg.EmbeddingTableName != "embeddings" {
		t.Fatalf("unexpected embedding table config: %+v", cfg)
	}
	if cfg.GeminiAPIKey != "gemini-key" || cfg.AuthSecretKey != "jwt-secret" || cfg.LogFile != "app.log" {
		t.Fatalf("unexpected secret/log config: %+v", cfg)
	}
}
