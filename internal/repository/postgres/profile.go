package postgres

import (
	"database/sql"
	"errors"
	"log"

	"conduit/internal/repository"

	"github.com/lib/pq"
)

// profileRepository implements the repository.profileRepository using PostgreSQL
type profileRepository struct {
	db *sql.DB
}

// NewProfileRepository creates a new profile repository
func NewProfileRepository(db *sql.DB) *profileRepository {
	return &profileRepository{db: db}
}

// GetByUsername gets a profile by username
func (r *profileRepository) GetByUsername(
	username string,
	currentUserID int64,
) (*repository.Profile, error) {
	query := `
		SELECT id, username, bio, image, following
		FROM profiles
		WHERE username = $1
	`

	var profile repository.Profile
	err := r.db.QueryRow(query, username).
		Scan(&profile.ID, &profile.Username, &profile.Bio, &profile.Image, &profile.Following)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	if currentUserID != 0 {
		var following bool
		err := r.db.QueryRow(
			`SELECT EXISTS (
				SELECT 1
				FROM follows
				WHERE follower_id = $1 AND following_id = $2
			)`, currentUserID, profile.ID).Scan(&following)
		if err != nil {
			return nil, repository.ErrInternal
		}

		profile.Following = following
	}

	return &profile, nil
}

// FollowUser follows a user
func (r *profileRepository) FollowUser(
	followerID int64,
	followingName string,
) (*repository.Profile, error) {
	// Begin a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("transaction rollback error: %v", err)
		}
	}()

	query := `
		WITH following_user AS (
			SELECT id
			FROM users
			WHERE username = $1
		),
		follow_attempt AS (
			INSERT INTO follows (follower_id, following_id)
			SELECT $2, id FROM following_user
			ON CONFLICT (follower_id, following_id) DO NOTHING
		)
		SELECT u.id, u.username, u.bio, u.image, true AS following
		FROM users u
		JOIN following_user fu ON u.id = fu.id
	`

	var profile repository.Profile
	var bio, image sql.NullString

	err = tx.QueryRow(query, followingName, followerID).
		Scan(&profile.ID, &profile.Username, &bio, &image, &profile.Following)
	if err != nil {
		// following_user does not exist
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		// PostgreSQL specific error handling
		if pqErr, ok := err.(*pq.Error); ok {
			// Foreign key constraint violation
			if pqErr.Code == "23503" {
				if pqErr.Constraint == "follows_follower_id_fkey" {
					return nil, repository.ErrUserNotFound
				}
			}
			// Check for self-follow constraint violation
			if pqErr.Code == "23514" {
				if pqErr.Constraint == "prevent_self_follow" {
					return nil, repository.ErrCannotFollowSelf
				}
			}
		}

		return nil, repository.ErrInternal
	}

	// Handle nullable values
	if bio.Valid {
		profile.Bio = bio.String
	}
	if image.Valid {
		profile.Image = image.String
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	return &profile, nil
}
