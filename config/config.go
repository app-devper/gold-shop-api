package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// ErrJWTSecretRequired is returned when JWT_SECRET is not set in production
var ErrJWTSecretRequired = errors.New("JWT_SECRET environment variable is required in production")

type Config struct {
	Server  ServerConfig
	MongoDB MongoDBConfig
	JWT     JWTConfig
	GoldAPI GoldAPIConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type MongoDBConfig struct {
	URI      string
	Database string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type GoldAPIConfig struct {
	URL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	env := getEnv("SERVER_ENV", "development")
	jwtSecret := os.Getenv("JWT_SECRET")

	// In production, JWT_SECRET must be explicitly set
	if env == "production" && jwtSecret == "" {
		return nil, ErrJWTSecretRequired
	}

	// In development, use a default secret with warning
	if jwtSecret == "" {
		jwtSecret = "dev-secret-key-do-not-use-in-production"
	}

	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  env,
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "gold_shop"),
		},
		JWT: JWTConfig{
			Secret:          jwtSecret,
			ExpirationHours: jwtExpHours,
		},
		GoldAPI: GoldAPIConfig{
			URL: getEnv("GOLD_API_URL", "https://api.chnwt.dev/thai-gold-api/latest"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
