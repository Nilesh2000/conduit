package postgres

import (
	"context"
	"database/sql"
)

// tagRepository implements the TagRepository interface
type tagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *sql.DB) *tagRepository {
	return &tagRepository{db: db}
}

// GetTags gets all tags
func (r *tagRepository) Get(ctx context.Context) ([]string, error) {
	query := `
		SELECT name FROM tags
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}
