package config

import (
	"log"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	HTTPPort     string   `env:"HTTP_PORT" envDefault:"8080"`
	DatabaseURL  string   `env:"DATABASE_URL" envDefault:"postgres://user:password@localhost:5432/orders_db"`
	KafkaBrokers []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	KafkaTopic   string   `env:"KAFKA_TOPIC" envDefault:"orders"`
}

func Load() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v", err)
	}
	return &cfg
}
