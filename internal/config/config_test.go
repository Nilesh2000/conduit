package config

import (
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
