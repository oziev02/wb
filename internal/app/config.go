package app

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DBURL             string        `env:"DB_URL,required"`
	HTTPAddr          string        `env:"HTTP_ADDR" envDefault:":8081"`
	KafkaBrokers      []string      `env:"KAFKA_BROKERS" envSeparator:","`
	KafkaTopic        string        `env:"KAFKA_TOPIC" envDefault:"orders"`
	KafkaGroup        string        `env:"KAFKA_GROUP" envDefault:"orders-consumer"`
	CacheCap          int           `env:"CACHE_CAP" envDefault:"10000"`
	CacheTTL          time.Duration `env:"CACHE_TTL" envDefault:"30m"`
	CacheRestoreLimit int           `env:"CACHE_RESTORE_LIMIT" envDefault:"10000"`
}

func LoadConfig() (Config, error) {
	var c Config
	err := env.Parse(&c)
	return c, err
}
