package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port              string
	AllowedOrigins    []string
	MongoURI          string
	JWTSecret         []byte
	JWTRefreshSecret  []byte
	JWTExpiry         time.Duration
	RefreshExpiry     time.Duration
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	Auth0RedirectURI  string
	KeyWrappingKey    string
	APIKeyHashSecret  string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8000"),
		AllowedOrigins:    getEnvSlice("ALLOWED_ORIGINS", []string{"http://lockkeep.localhost"}),
		MongoURI:          getEnv("MONGODB_URI", "mongodb://localhost:27017/lockkeep"),
		JWTSecret:         []byte(getEnv("JWT_SECRET", "change-me-in-production-min-32-characters-long")),
		JWTRefreshSecret:  []byte(getEnv("JWT_REFRESH_SECRET", "different-change-me-in-production")),
		JWTExpiry:         5 * time.Minute,
		RefreshExpiry:     7 * 24 * time.Hour,
		Auth0Domain:       getEnv("AUTH0_DOMAIN", ""),
		Auth0ClientID:     getEnv("AUTH0_CLIENT_ID", ""),
		Auth0ClientSecret: getEnv("AUTH0_CLIENT_SECRET", ""),
		Auth0RedirectURI:  getEnv("AUTH0_REDIRECT_URI", "http://localhost:5173/callback"),
		KeyWrappingKey:    getEnv("LOCKKEEP_KEK", ""),
		APIKeyHashSecret:  getEnv("LOCKKEEP_API_KEY_HASH_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); len(v) > 0 {
		return strings.Split(v, ",")
	}
	return fallback
}
