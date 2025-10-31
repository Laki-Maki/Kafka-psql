package config

import (
	"log"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
	Cache    CacheConfig
}

type ServerConfig struct { Port int }
type PostgresConfig struct {
	Host, User, Password, DBName, SSLMode string
	Port int
}
type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}
type CacheConfig struct { Size int }

func LoadConfig(path string) Config {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}
	return cfg
}
