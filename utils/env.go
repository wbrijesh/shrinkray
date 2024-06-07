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

func VerifyEnv(name string) bool {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if os.Getenv(name) == "" {
		return false
	}
	return true
}
