package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func InitEnv() {
	_ = godotenv.Load()

	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",

		"HTTP_ADDRESS",
		"HTTP_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
	}

	for _, key := range required {
		val, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("missing required env variable: %s", key)
		}

		if key == "DB_PORT" || key == "HTTP_TIMEOUT" || key == "HTTP_IDLE_TIMEOUT" {
			if _, err := strconv.Atoi(val); err != nil {
				log.Fatalf("invalid integer env variable: %s=%s", key, val)
			}
		}
	}
}
