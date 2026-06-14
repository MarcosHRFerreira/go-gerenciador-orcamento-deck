package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	SecretJWT   string
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		SecretJWT:   strings.TrimSpace(os.Getenv("SECRET_JWT")),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DBHost:      strings.TrimSpace(os.Getenv("DB_HOST")),
		DBPort:      getEnvOrDefault("DB_PORT", "5432"),
		DBUser:      strings.TrimSpace(os.Getenv("DB_USER")),
		DBPassword:  strings.TrimSpace(os.Getenv("DB_PASSWORD")),
		DBName:      strings.TrimSpace(os.Getenv("DB_NAME")),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) ServerAddress() string {
	return ":" + c.Port
}

func (c *Config) validate() error {
	missingFields := make([]string, 0, 6)

	if c.SecretJWT == "" {
		missingFields = append(missingFields, "SECRET_JWT")
	}
	if c.DatabaseURL == "" {
		missingFields = append(missingFields, "DATABASE_URL")
	}
	if c.DBHost == "" {
		missingFields = append(missingFields, "DB_HOST")
	}
	if c.DBUser == "" {
		missingFields = append(missingFields, "DB_USER")
	}
	if c.DBPassword == "" {
		missingFields = append(missingFields, "DB_PASSWORD")
	}
	if c.DBName == "" {
		missingFields = append(missingFields, "DB_NAME")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}
