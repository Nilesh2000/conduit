package repository

import "time"

// User represents a user in the repository
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	Bio          string
	Image        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
