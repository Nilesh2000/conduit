package postgres

import (
	"conduit/internal/repository"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// userRepository implements the repository.userRepository using PostgreSQL
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

// Create creates a new user in the database
func (r *userRepository) Create(username, email, password string) (*repository.User, error) {
	// Begin a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, password
	`

	now := time.Now()
	var user repository.User

	// Insert user
	err = tx.QueryRow(query, username, email, password, now, now).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
	)

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
	return &user, nil
}

// FindByEmail finds a user by email in the database
func (r *userRepository) FindByEmail(email string) (*repository.User, error) {
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
func (r *userRepository) FindByID(id int64) (*repository.User, error) {
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

// Update updates a user in the database
func (r *userRepository) Update(userID int64, username, email, password, bio, image *string) (*repository.User, error) {
	// Begin a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer tx.Rollback()

	query := `
		UPDATE users
		SET
			username = COALESCE($1, username),
			email = COALESCE($2, email),
			password = COALESCE($3, password),
			bio = COALESCE($4, bio),
			image = COALESCE($5, image),
			updated_at = $6
		WHERE id = $7
		RETURNING id, username, email, password, bio, image
	`

	now := time.Now()
	var updatedUser repository.User
	var nsBio, nsImage sql.NullString

	// Update user
	err = tx.QueryRow(
		query,
		username,
		email,
		password,
		bio,
		image,
		now,
		userID,
	).Scan(&updatedUser.ID, &updatedUser.Username, &updatedUser.Email, &updatedUser.Password, &nsBio, &nsImage)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}

		// PostgreSQL specific error handling
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

	// Handle nullable values
	if nsBio.Valid {
		updatedUser.Bio = nsBio.String
	}

	if nsImage.Valid {
		updatedUser.Image = nsImage.String
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	return &updatedUser, nil
}
