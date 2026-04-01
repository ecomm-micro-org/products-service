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
	DSN                string
	Brokers            []string
	CacheAddr          string
	CachePasswd        string
	Port               string
	ServiceRegistry    string
	EurekaHostname     string
	EmbeddingModelURL  string
	EmbeddingModelName string
}

func Config() *config {
	return defaultConfig
}

func Init() {
	once.Do(func() {
		brokers := os.Getenv("BROKERS")
		defaultConfig = &config{
			DSN:                os.Getenv("DSN"),
			Brokers:            strings.Split(brokers, ","),
			CacheAddr:          os.Getenv("CACHE_ADDR"),
			CachePasswd:        os.Getenv("CACHE_PASSWD"),
			Port:               os.Getenv("PORT"),
			EurekaHostname:     os.Getenv("EUREKA_HOSTNAME"),
			ServiceRegistry:    os.Getenv("SERVICE_REGISTRY"),
			EmbeddingModelURL:  os.Getenv("EMBEDDING_MODEL_URL"),
			EmbeddingModelName: os.Getenv("EMBEDDING_MODEL_NAME"),
		}
	})
}
