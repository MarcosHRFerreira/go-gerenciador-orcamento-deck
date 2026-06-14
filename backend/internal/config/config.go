package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	AppEnv                 string
	LogLevel               string
	SecretJWT              string
	DatabaseURL            string
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	AllowedOrigins         []string
	InitialAdminSetupToken string
	RefreshCookieName      string
	RefreshCookieDomain    string
	RefreshCookieSecure    bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                   getEnvOrDefault("PORT", "8080"),
		AppEnv:                 getEnvOrDefault("APP_ENV", "development"),
		LogLevel:               getEnvOrDefault("LOG_LEVEL", "info"),
		SecretJWT:              strings.TrimSpace(os.Getenv("SECRET_JWT")),
		DatabaseURL:            strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DBHost:                 strings.TrimSpace(os.Getenv("DB_HOST")),
		DBPort:                 getEnvOrDefault("DB_PORT", "5432"),
		DBUser:                 strings.TrimSpace(os.Getenv("DB_USER")),
		DBPassword:             strings.TrimSpace(os.Getenv("DB_PASSWORD")),
		DBName:                 strings.TrimSpace(os.Getenv("DB_NAME")),
		AllowedOrigins:         getCSVEnvOrDefault("ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		InitialAdminSetupToken: strings.TrimSpace(os.Getenv("INITIAL_ADMIN_SETUP_TOKEN")),
		RefreshCookieName:      getEnvOrDefault("REFRESH_COOKIE_NAME", "budget_management_refresh"),
		RefreshCookieDomain:    strings.TrimSpace(os.Getenv("REFRESH_COOKIE_DOMAIN")),
		RefreshCookieSecure:    getBoolEnvOrDefault("REFRESH_COOKIE_SECURE", false),
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
	if len(c.AllowedOrigins) == 0 {
		missingFields = append(missingFields, "ALLOWED_ORIGINS")
	}
	if c.RefreshCookieName == "" {
		missingFields = append(missingFields, "REFRESH_COOKIE_NAME")
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

func getCSVEnvOrDefault(key string, defaultValues []string) []string {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return defaultValues
	}

	items := strings.Split(rawValue, ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		normalizedItem := strings.TrimSpace(item)
		if normalizedItem != "" {
			values = append(values, normalizedItem)
		}
	}

	if len(values) == 0 {
		return defaultValues
	}

	return values
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	normalizedValue := strings.ToLower(value)
	return normalizedValue == "1" || normalizedValue == "true" || normalizedValue == "yes"
}
