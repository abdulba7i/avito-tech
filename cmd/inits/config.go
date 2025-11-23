package inits

import (
	"os"
	"strconv"
)

type Config struct {
	Postgres   PostgresConfig
	HttpServer HttpServer
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type HttpServer struct {
	Address     string
	Timeout     int
	IdleTimeout int
}

func LoadConfig() (*Config, error) {
	timeout, _ := strconv.Atoi(os.Getenv("HTTP_TIMEOUT"))
	idleTimeout, _ := strconv.Atoi(os.Getenv("HTTP_IDLE_TIMEOUT"))

	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))

	cfg := &Config{
		HttpServer: HttpServer{
			Address:     os.Getenv("HTTP_ADDRESS"),
			Timeout:     timeout,
			IdleTimeout: idleTimeout,
		},
		Postgres: PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     port,
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
	}

	return cfg, nil
}
