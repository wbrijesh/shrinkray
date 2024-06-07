package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv(name string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv(name)
}
