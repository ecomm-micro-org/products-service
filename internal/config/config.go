package config

import (
	"os"
	"strings"
	"sync"
)

var (
	defaultConfig *config
	once          sync.Once
)

type config struct {
	DSN                          string
	Brokers                      []string
	CacheAddr                    string
	CachePasswd                  string
	Port                         string
	ServiceRegistry              string
	EurekaHostname               string
	EmbeddingCollectionTableName string
	EmbeddingTableName           string
	GeminiAPIKey                 string
	AuthSecretKey                string
	LogFile                      string
}

func Config() *config {
	return defaultConfig
}

func Init() {
	once.Do(func() {
		brokers := os.Getenv("BROKERS")
		defaultConfig = &config{
			DSN:                          os.Getenv("DSN"),
			Brokers:                      strings.Split(brokers, ","),
			CacheAddr:                    os.Getenv("CACHE_ADDR"),
			CachePasswd:                  os.Getenv("CACHE_PASSWD"),
			Port:                         os.Getenv("PORT"),
			EurekaHostname:               os.Getenv("EUREKA_HOSTNAME"),
			ServiceRegistry:              os.Getenv("SERVICE_REGISTRY"),
			EmbeddingCollectionTableName: os.Getenv("EMBEDDING_COLLECTION_TABLE_NAME"),
			EmbeddingTableName:           os.Getenv("EMBEDDING_TABLE_NAME"),
			GeminiAPIKey:                 os.Getenv("GEMINI_API_KEY"),
			AuthSecretKey:                os.Getenv("SECRET_KEY"),
			LogFile:                      os.Getenv("LOG_FILE"),
		}
	})
}
