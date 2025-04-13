package postgres

import (
	"conduit/internal/repository"
	"database/sql"
	"errors"
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
func (r *profileRepository) GetByUsername(username string, currentUserID int64) (*repository.Profile, error) {
	query := `
		SELECT id, username, bio, image, following
		FROM profiles
		WHERE username = $1
	`

	var profile repository.Profile
	err := r.db.QueryRow(query, username).Scan(&profile.ID, &profile.Username, &profile.Bio, &profile.Image, &profile.Following)
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
