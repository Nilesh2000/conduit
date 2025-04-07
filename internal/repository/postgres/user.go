package postgres

import (
	"conduit/internal/repository"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username, email, password string) (*repository.User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer tx.Rollback()

	now := time.Now()
	var userID int64
	err = tx.QueryRow("INSERT INTO users (username, email, password, bio, image, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", username, email, password, "", "", now, now).Scan(&userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				if pqErr.Constraint == "users_username_key" {
					return nil, repository.ErrDuplicateUsername
				}
				if pqErr.Constraint == "users_email_key" {
					return nil, repository.ErrDuplicateEmail
				}
			}
		}
		return nil, repository.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	return &repository.User{
		ID:       userID,
		Username: username,
		Email:    email,
		Password: password,
		Bio:      "",
		Image:    "",
	}, nil
}
