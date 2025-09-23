package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN             string
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshSecret     string
	RefreshExpiration time.Duration
	Port              string
	Env               string
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	requiredEnvVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"JWT_SECRET", "REFRESH_SECRET", "PORT", "ENV",
	}
	for _, env := range requiredEnvVars {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf("falta la variable de entorno requerida: %s", env)
		}
	}

	jwtExp, err := parseDuration(os.Getenv("JWT_EXPIRATION"), 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION inválido: %w", err)
	}

	refreshExp, err := parseDuration(os.Getenv("REFRESH_EXPIRATION"), 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("REFRESH_EXPIRATION inválido: %w", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	return &Config{
		DBDSN:             dsn,
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiration:     jwtExp,
		RefreshSecret:     os.Getenv("REFRESH_SECRET"),
		RefreshExpiration: refreshExp,
		Port:              os.Getenv("PORT"),
		Env:               os.Getenv("ENV"),
	}, nil
}

func parseDuration(value string, defaultDuration time.Duration) (time.Duration, error) {
	if value == "" {
		return defaultDuration, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultDuration, fmt.Errorf("error al parsear duración: %w", err)
	}
	return d, nil
}
