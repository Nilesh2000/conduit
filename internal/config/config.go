package config

import (
	"fmt"
	"os"
	"time"
)

// Config represents the application configuration.
type Config struct {
	Database Database
	JWT      JWT
	Server   Server
}

// Database represents the database configuration.
type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWT represents the JWT configuration.
type JWT struct {
	SecretKey string
	Expiry    time.Duration
}

// Server represents the server configuration.
type Server struct {
	Port string
}

// Load loads the configuration from the environment variables.
func Load() (*Config, error) {
	jwtExpiry := getEnv("JWT_EXPIRY", "24h")
	expiry, err := time.ParseDuration(jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT expiry duration: %w", err)
	}

	cfg := &Config{
		Database: Database{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "admin"),
			Name:     getEnv("DB_NAME", "conduit"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWT{
			SecretKey: getEnv("JWT_SECRET_KEY", "secret-key"),
			Expiry:    expiry,
		},
		Server: Server{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}

	return cfg, nil
}

// GetDSN returns the database connection string.
func (c *Database) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}
	if c.JWT.Expiry <= 0 {
		return fmt.Errorf("JWT expiry is required")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	return nil
}

// getEnv returns the value of the environment variable.
// If the variable is not set, it returns the default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
