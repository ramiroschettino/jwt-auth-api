package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN             string
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env vars")
	}

	jwtExpStr := os.Getenv("JWT_EXPIRATION")
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		jwtExp = 15 * time.Minute
		log.Printf("Invalid JWT_EXPIRATION, using default: %v", jwtExp)
	}

	refreshExpStr := os.Getenv("REFRESH_EXPIRATION")
	refreshExp, err := time.ParseDuration(refreshExpStr)
	if err != nil {
		refreshExp = 24 * time.Hour
		log.Printf("Invalid REFRESH_EXPIRATION, using default: %v", refreshExp)
	}

	return &Config{
		DBDSN:             os.Getenv("DB_DSN"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiration:     jwtExp,
		RefreshExpiration: refreshExp,
	}
}
