package config

import (
	"fmt"
	"os"
	"strconv"
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
			SecretKey: getEnv("JWT_SECRET_KEY", "this-is-a-32-char-long-secret-key-123"),
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
	// Validate database configuration
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database configuration error: %w", err)
	}

	// Validate JWT configuration
	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("JWT configuration error: %w", err)
	}

	// Validate server configuration
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server configuration error: %w", err)
	}

	return nil
}

// Validate checks if the database configuration is valid.
func (d *Database) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("host is required")
	}
	if d.Port == "" {
		return fmt.Errorf("port is required")
	}
	if d.User == "" {
		return fmt.Errorf("user is required")
	}
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate port is a number
	if _, err := strconv.Atoi(d.Port); err != nil {
		return fmt.Errorf("port must be a valid number: %w", err)
	}

	return nil
}

// Validate checks if the JWT configuration is valid.
func (j *JWT) Validate() error {
	if j.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if j.Expiry <= 0 {
		return fmt.Errorf("expiry must be greater than 0")
	}

	// Validate secret key is at least 32 bytes long for security
	if len(j.SecretKey) < 32 {
		return fmt.Errorf("secret key must be at least 32 bytes long for security")
	}

	return nil
}

// Validate checks if the server configuration is valid.
func (s *Server) Validate() error {
	if s.Port == "" {
		return fmt.Errorf("port is required")
	}

	// Validate port is a number and in valid range
	port, err := strconv.Atoi(s.Port)
	if err != nil {
		return fmt.Errorf("port must be a valid number: %w", err)
	}
	if port < 0 || port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
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
