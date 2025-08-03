package config

import (
	"os"
	"testing"
	"time"
)

func TestDatabase_GetDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		database Database
		expected string
	}{
		{
			name: "Standard DSN",
			database: Database{
				Host:     "testhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			expected: "postgres://testuser:testpass@testhost:5432/testdb?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.database.GetDSN(); got != tt.expected {
				t.Errorf("GetDSN() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: Config{
				Database: Database{
					Host:     "localhost",
					Port:     "5432",
					User:     "testuser",
					Password: "testpass",
					Name:     "testdb",
					SSLMode:  "disable",

					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 10 * time.Second,
					ConnMaxIdleTime: 5 * time.Second,
				},
				JWT: JWT{
					SecretKey: "this-is-a-32-char-long-secret-key-123",
					Expiry:    24 * time.Hour,
				},
				Server: Server{
					Port: "8080",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing database host",
			config: Config{
				Database: Database{
					Port:     "5432",
					User:     "testuser",
					Password: "testpass",
					Name:     "testdb",
					SSLMode:  "disable",

					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 10 * time.Second,
					ConnMaxIdleTime: 5 * time.Second,
				},
				JWT: JWT{
					SecretKey: "this-is-a-32-char-long-secret-key-123",
					Expiry:    24 * time.Hour,
				},
				Server: Server{
					Port: "8080",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid JWT expiry",
			config: Config{
				Database: Database{
					Host:     "localhost",
					Port:     "5432",
					User:     "testuser",
					Password: "testpass",
					Name:     "testdb",
					SSLMode:  "disable",

					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 10 * time.Second,
					ConnMaxIdleTime: 5 * time.Second,
				},
				JWT: JWT{
					SecretKey: "this-is-a-32-char-long-secret-key-123",
					Expiry:    -1 * time.Hour,
				},
				Server: Server{
					Port: "8080",
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Server Port",
			config: Config{
				Database: Database{
					Host:     "localhost",
					Port:     "5432",
					User:     "testuser",
					Password: "testpass",
					Name:     "testdb",
					SSLMode:  "disable",

					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 10 * time.Second,
					ConnMaxIdleTime: 5 * time.Second,
				},
				JWT: JWT{
					SecretKey: "this-is-a-32-char-long-secret-key-123",
					Expiry:    24 * time.Hour,
				},
				Server: Server{
					Port: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_WithEnvFile(t *testing.T) {
	t.Parallel()

	// Create a temporary .env file
	envContent := `DB_HOST=testhost
DB_PORT=5433
DB_USER=testuser
DB_PASSWORD=testpass
DB_NAME=testdb
DB_SSLMODE=require
JWT_SECRET_KEY=test-secret-key-that-is-long-enough-for-security
JWT_EXPIRY=12h
SERVER_PORT=9090
APP_VERSION=2.0.0`

	// Write temporary .env file
	tmpFile, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(envContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Change to the directory with the .env file
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	// Change to temp directory
	if err := os.Chdir(os.TempDir()); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Rename the temp file to .env in the temp directory
	envPath := os.TempDir() + "/.env"
	if err := os.Rename(tmpFile.Name(), envPath); err != nil {
		t.Fatalf("Failed to rename temp file: %v", err)
	}
	defer os.Remove(envPath)

	// Load configuration
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify the configuration was loaded from .env file
	if cfg.Database.Host != "testhost" {
		t.Errorf("Expected DB_HOST to be 'testhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != "5433" {
		t.Errorf("Expected DB_PORT to be '5433', got '%s'", cfg.Database.Port)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("Expected DB_USER to be 'testuser', got '%s'", cfg.Database.User)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("Expected DB_PASSWORD to be 'testpass', got '%s'", cfg.Database.Password)
	}
	if cfg.Database.Name != "testdb" {
		t.Errorf("Expected DB_NAME to be 'testdb', got '%s'", cfg.Database.Name)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Expected DB_SSLMODE to be 'require', got '%s'", cfg.Database.SSLMode)
	}
	if cfg.JWT.SecretKey != "test-secret-key-that-is-long-enough-for-security" {
		t.Errorf(
			"Expected JWT_SECRET_KEY to be 'test-secret-key-that-is-long-enough-for-security', got '%s'",
			cfg.JWT.SecretKey,
		)
	}
	if cfg.JWT.Expiry != 12*time.Hour {
		t.Errorf("Expected JWT_EXPIRY to be 12h, got '%v'", cfg.JWT.Expiry)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("Expected SERVER_PORT to be '9090', got '%s'", cfg.Server.Port)
	}
	if cfg.Version != "2.0.0" {
		t.Errorf("Expected APP_VERSION to be '2.0.0', got '%s'", cfg.Version)
	}
}
