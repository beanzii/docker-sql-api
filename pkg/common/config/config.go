package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Server struct {
	Port uint16 `env:"SERVER_PORT"`
}

type Database struct {
	Port     uint16 `env:"DB_PORT"`
	Host     string `env:"DB_HOST"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	Driver   string `env:"DB_DRIVER"`
}

type Configs interface {
	Server | Database
}

func Load(envLocation string) error {
	err := godotenv.Load(envLocation)
	if err != nil {
		log.Printf("[ERROR] Unable to Load() .env file from %s", envLocation)
	}
	return err
}

func Parse[T Configs](config T) (T, error) {
	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing %T config: %s", config, err)
	}
	return config, err
}
