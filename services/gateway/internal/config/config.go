package config

import (
	"errors"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr string `env:"HTTP_PORT" envDefault:":8080"`

	UserGRPCAddr    string `env:"USER_GRPC_ADDR" envDefault:"user:50051"`
	ProductGRPCAddr string `env:"PRODUCT_GRPC_ADDR" envDefault:"product:50052"`
	OrderGRPCAddr   string `env:"ORDER_GRPC_ADDR" envDefault:"order:50053"`

	AccessSecretKey  string `env:"ACCESS_SECRET_KEY,required"`
	RefreshSecretKey string `env:"REFRESH_SECRET_KEY,required"`
}

func MustConfig(envPath string) *Config {
	err := godotenv.Load(envPath)
	if err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		log.Fatalf("load .env: %v", err)
	}

	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("parse config: %v", err)
	}

	return &cfg
}
