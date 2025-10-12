package config

import (
	"os"
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

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "civicweave"),
			User:     getEnv("DB_USER", "civicweave"),
			Password: getEnv("DB_PASSWORD", "civicweave_dev"),
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
	}
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
