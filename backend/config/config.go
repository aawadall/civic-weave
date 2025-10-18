package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Mailgun   MailgunConfig
	Google    GoogleConfig
	Geocoding GeocodingConfig
	OpenAI    OpenAIConfig
	Features  FeatureFlags
	CORS      CORSConfig
}

// FeatureFlags holds feature toggle settings
type FeatureFlags struct {
	EmailEnabled bool
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret string
}

// MailgunConfig holds Mailgun settings
type MailgunConfig struct {
	APIKey string
	Domain string
}

// GoogleConfig holds Google OAuth settings
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
}

// GeocodingConfig holds geocoding service settings
type GeocodingConfig struct {
	NominatimBaseURL string
}

// OpenAIConfig holds OpenAI API settings
type OpenAIConfig struct {
	APIKey         string
	EmbeddingModel string
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "civicweave"),
			User:     getEnv("DB_USER", "civicweave"),
			Password: getEnv("DB_PASSWORD", "civicweave_dev"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "dev_jwt_secret_key_change_in_production"),
		},
		Mailgun: MailgunConfig{
			APIKey: getEnv("MAILGUN_API_KEY", ""),
			Domain: getEnv("MAILGUN_DOMAIN", ""),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
		Geocoding: GeocodingConfig{
			NominatimBaseURL: getEnv("NOMINATIM_BASE_URL", "https://nominatim.openstreetmap.org"),
		},
		OpenAI: OpenAIConfig{
			APIKey:         getEnv("OPENAI_API_KEY", ""),
			EmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		},
		Features: FeatureFlags{
			EmailEnabled: getEnv("ENABLE_EMAIL", "true") == "true",
		},
		CORS: CORSConfig{
			AllowedOrigins: parseCORSOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,https://civicweave.com,https://civicweave-frontend-162941711179.us-central1.run.app")),
		},
	}
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseCORSOrigins parses comma-separated CORS origins
func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}
	
	var result []string
	for _, origin := range strings.Split(origins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
