// Package config loads service configuration from APP_* environment variables.
package config

import (
	"fmt"
	"strings"
)

// Config is the process configuration.
type Config struct {
	// Name is the service's name: the events' source, the logs' service field.
	Name string
	// Addr is the listener: Connect, gRPC, gRPC-Web and REST on one port.
	Addr string
	// DatabaseURL is the Postgres primary: every write and migrations.
	DatabaseURL string
	// DatabaseReadURL is a Postgres read replica for reads outside a unit of work (defaults to
	// DatabaseURL). Replica lag is accepted: a read right after a write may miss it.
	DatabaseReadURL string
	// RedisURL is the Redis the cache uses.
	RedisURL string
	// KafkaBrokers and KafkaTopic are where events go; without brokers, events are only logged.
	KafkaBrokers []string
	KafkaTopic   string
}

// Load reads configuration through getenv (pass os.Getenv in production).
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Name:            orDefault(getenv("APP_NAME"), "base"),
		Addr:            orDefault(getenv("APP_ADDR"), ":8080"),
		DatabaseURL:     getenv("APP_DATABASE_URL"),
		DatabaseReadURL: getenv("APP_DATABASE_READ_URL"),
		RedisURL:        getenv("APP_REDIS_URL"),
	}
	cfg.KafkaTopic = orDefault(getenv("APP_KAFKA_TOPIC"), cfg.Name+".events")
	for _, b := range strings.Split(getenv("APP_KAFKA_BROKERS"), ",") {
		if b = strings.TrimSpace(b); b != "" {
			cfg.KafkaBrokers = append(cfg.KafkaBrokers, b)
		}
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("APP_DATABASE_URL is required")
	}
	if cfg.DatabaseReadURL == "" {
		cfg.DatabaseReadURL = cfg.DatabaseURL
	}
	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("APP_REDIS_URL is required")
	}
	return cfg, nil
}

func orDefault(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
