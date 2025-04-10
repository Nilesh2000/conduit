package postgres

import (
	"conduit/internal/repository"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// UserRepository implements the repository.UserRepository using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// New creates a new UserRepository
func New(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(username, email, password string) (*repository.User, error) {
	// Begin a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer tx.Rollback()

	// Insert user
	now := time.Now()
	var userID int64
	err = tx.QueryRow("INSERT INTO users (username, email, password, bio, image, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", username, email, password, "", "", now, now).Scan(&userID)

	// Handle database errors
	if err != nil {
		// PostgreSQL specific error handling
		if pqErr, ok := err.(*pq.Error); ok {
			// Check for unique violation error codes
			if pqErr.Code == "23505" {
				// Determine which constraint was violated
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

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	// Return the created user
	return &repository.User{
		ID:       userID,
		Username: username,
		Email:    email,
		Password: password,
		Bio:      "",
		Image:    "",
	}, nil
}

// FindByEmail finds a user by email in the database
func (r *UserRepository) FindByEmail(email string) (*repository.User, error) {
	var user repository.User
	var bio, image sql.NullString

	err := r.db.QueryRow("SELECT id, username, email, password, bio, image FROM users WHERE email = $1", email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &bio, &image)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, repository.ErrInternal
	}

	// Handle nullable values
	if bio.Valid {
		user.Bio = bio.String
	}

	if image.Valid {
		user.Image = image.String
	}

	return &user, nil
}

// FindByID finds a user by ID in the database
func (r *UserRepository) FindByID(id int64) (*repository.User, error) {
	var user repository.User
	var bio, image sql.NullString

	err := r.db.QueryRow("SELECT id, username, email, password, bio, image FROM users WHERE id = $1", id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &bio, &image)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, repository.ErrInternal
	}

	// Handle nullable values
	if bio.Valid {
		user.Bio = bio.String
	}

	if image.Valid {
		user.Image = image.String
	}

	return &user, nil
}
