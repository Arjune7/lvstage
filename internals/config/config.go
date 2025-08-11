package config

import (
	"lystage-proj/internals/observability"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	GinMode     string
	KafkaBroker string
	ClicksTopic string
}

func Load() *Config {
	// Load .env if present (only for local/dev)
	_ = godotenv.Load(".env")

	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Port:        getEnv("PORT", "5000"),
		GinMode:     getEnv("GIN_MODE", "release"),
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		ClicksTopic: getEnv("CLICKS_TOPIC", "click-events"),
	}
	observability.Logger.Info(cfg.DatabaseURL)
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
