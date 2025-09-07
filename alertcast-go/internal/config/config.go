package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	KafkaBroker  string
	TopicEvents  string
	ConsumerGroup string

	PGHost string
	PGPort string
	PGUser string
	PGPassword string
	PGDB   string

	RedisAddr string

	APIPort string

	IngestRate float64
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func New() (*Config, error) {
	rateStr := getenv("INGEST_RATE", "0.2")
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid INGEST_RATE: %w", err)
	}

	return &Config{
		KafkaBroker:  getenv("KAFKA_BROKER", "kafka:9092"),
		TopicEvents:  getenv("TOPIC_EVENTS", "device_events"),
		ConsumerGroup: getenv("CONSUMER_GROUP", "alertcast-processor"),

		PGHost: getenv("PG_HOST", "postgres"),
		PGPort: getenv("PG_PORT", "5432"),
		PGUser: getenv("PG_USER", "alertcast"),
		PGPassword: getenv("PG_PASSWORD", "alertcast"),
		PGDB:   getenv("PG_DB", "alertcast"),

		RedisAddr: getenv("REDIS_ADDR", "redis:6379"),
		APIPort:   getenv("API_PORT", "8080"),
		IngestRate: rate,
	}, nil
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.PGUser, c.PGPassword, c.PGHost, c.PGPort, c.PGDB)
}
