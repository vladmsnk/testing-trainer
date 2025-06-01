package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	pathToDevEnv    = "./etc/dev.env"
	pathToConfigEnv = "./etc/config.env"
)

type Config struct {
	HTTP Http
	PG   Postgres
}

type Http struct {
	Host string
	Port int
}

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func (p *Postgres) GetConnectionString() string {
	return "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + strconv.Itoa(p.Port) + "/" + p.Database + "?" + "sslmode" + "=" + p.SSLMode
}

var ConfigStruct Config

// getEnvOrDefault returns the environment variable value if set, otherwise returns the default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitConfigWithEnvs() error {
	// Determine which env file to use based on ENV variable
	// ENV=production uses config.env, otherwise uses dev.env (default)
	envFile := pathToDevEnv
	if os.Getenv("ENV") == "production" {
		envFile = pathToConfigEnv
	}

	// Load the .env file if it exists (provides defaults)
	_ = godotenv.Load(envFile) // Ignore error if file doesn't exist

	// Get values prioritizing environment variables over file values
	pgHost := getEnvOrDefault("PG_HOST", "localhost")
	pgPortStr := getEnvOrDefault("PG_PORT", "5432")
	pgUser := getEnvOrDefault("PG_USER", "postgres")
	pgPassword := getEnvOrDefault("PG_PASSWORD", "postgres")
	pgDatabase := getEnvOrDefault("PG_DATABASE", "db")
	pgSSLMode := getEnvOrDefault("PG_SSLMODE", "disable")

	httpHost := getEnvOrDefault("HTTP_HOST", "0.0.0.0")
	httpPortStr := getEnvOrDefault("HTTP_PORT", "7001")

	pgPort, err := strconv.Atoi(pgPortStr)
	if err != nil {
		return fmt.Errorf("invalid PG_PORT: %w", err)
	}

	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		return fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	ConfigStruct = Config{
		HTTP: Http{
			Host: httpHost,
			Port: httpPort,
		},
		PG: Postgres{
			Host:     pgHost,
			Port:     pgPort,
			User:     pgUser,
			Password: pgPassword,
			Database: pgDatabase,
			SSLMode:  pgSSLMode,
		},
	}
	return nil
}
