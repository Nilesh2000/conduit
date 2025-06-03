package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"
)

// commentRepository implements the CommentRepository interface
type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *sql.DB) *commentRepository {
	return &commentRepository{db: db}
}

// GetByID gets a comment by ID
func (r *commentRepository) GetByID(
	ctx context.Context,
	commentID int64,
) (*repository.Comment, error) {
	query := `
		SELECT id, body, article_id, user_id, created_at, updated_at
		FROM comments
		WHERE id = $1
		`

	var comment repository.Comment

	err := r.db.QueryRowContext(ctx, query, commentID).
		Scan(
			&comment.ID,
			&comment.Body,
			&comment.Article.ID,
			&comment.Author.ID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrCommentNotFound
		}
		return nil, repository.ErrInternal
	}

	return &comment, nil
}

// Create creates a new comment
func (r *commentRepository) Create(
	ctx context.Context,
	userID, articleID int64,
	body string,
) (*repository.Comment, error) {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("transaction rollback error: %v", err)
		}
	}()

	query := `
		WITH inserted_comment AS (
			INSERT INTO comments (body, article_id, user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, body, article_id, user_id, created_at, updated_at
		)
		SELECT
			c.id, c.created_at, c.updated_at, c.body,
			u.id AS author_id, u.username AS author_username, u.bio AS author_bio, u.image AS author_image,
			a.id AS article_id, a.slug AS article_slug, a.title AS article_title,
			a.description AS article_description, a.body AS article_body
		FROM inserted_comment c
		JOIN users u ON u.id = c.user_id
		JOIN articles a ON a.id = c.article_id
	`

	now := time.Now()
	var comment repository.Comment
	var authorBio, authorImage sql.NullString
	comment.Author = repository.Profile{}
	comment.Article = repository.Article{}

	if err := tx.QueryRowContext(ctx, query, body, articleID, userID, now, now).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt, &comment.Body,
			&comment.Author.ID, &comment.Author.Username, &authorBio, &authorImage,
			&comment.Article.ID, &comment.Article.Slug, &comment.Article.Title, &comment.Article.Description, &comment.Article.Body); err != nil {
		return nil, err
	}

	if authorBio.Valid {
		comment.Author.Bio = authorBio.String
	}
	if authorImage.Valid {
		comment.Author.Image = authorImage.String
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	return &comment, nil
}

// Delete deletes a comment
func (r *commentRepository) Delete(ctx context.Context, commentID int64) error {
	query := `
		DELETE FROM comments
		WHERE id = $1
		`

	result, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		return repository.ErrInternal
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return repository.ErrInternal
	}

	if rowsAffected == 0 {
		return repository.ErrCommentNotFound
	}

	return nil
}
