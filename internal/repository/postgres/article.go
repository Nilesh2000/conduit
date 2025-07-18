package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"

	"github.com/lib/pq"
)

// articleRepository implements the repository.articleRepository using PostgreSQL
type articleRepository struct {
	db *sql.DB
}

// NewArticleRepository creates a new ArticleRepository
func NewArticleRepository(db *sql.DB) *articleRepository {
	return &articleRepository{db: db}
}

// Create creates a new article in the database
func (r *articleRepository) Create(
	ctx context.Context,
	userID int64,
	slug, title, description, body string,
	tagList []string,
) (*repository.Article, error) {
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
		WITH inserted_article AS (
			INSERT INTO articles (slug, title, description, body, author_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, slug, title, description, body, author_id, created_at, updated_at
		)
		SELECT
			a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at,
			u.id, u.username, u.bio, u.image
		FROM inserted_article a
		JOIN users u ON u.id = a.author_id
	`

	now := time.Now()
	var article repository.Article
	article.Author = &repository.User{}
	var authorBio, authorImage sql.NullString

	err = tx.QueryRowContext(ctx, query, slug, title, description, body, userID, now, now).
		Scan(
			&article.ID,
			&article.Slug,
			&article.Title,
			&article.Description,
			&article.Body,
			&article.AuthorID,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.Author.ID,
			&article.Author.Username,
			&authorBio,
			&authorImage,
		)
	if err != nil {
		// PostgreSQL specific error handling
		if pqErr, ok := err.(*pq.Error); ok {
			// Check for foreign key constraint violation
			if pqErr.Code == "23503" && pqErr.Constraint == "articles_author_id_fkey" {
				return nil, repository.ErrUserNotFound
			}
			// Check for duplicate slug
			if pqErr.Code == "23505" {
				if pqErr.Constraint == "articles_slug_key" {
					return nil, repository.ErrDuplicateSlug
				}
			}
		}
		return nil, repository.ErrInternal
	}

	// Handle nullable values
	if authorBio.Valid {
		article.Author.Bio = authorBio.String
	}
	if authorImage.Valid {
		article.Author.Image = authorImage.String
	}

	// Add tags if any
	if len(tagList) > 0 {
		for _, tag := range tagList {
			var tagID int64
			// Insert or update tag and get its ID
			err := tx.QueryRowContext(ctx, "INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id", tag).
				Scan(&tagID)
			if err != nil {
				return nil, repository.ErrInternal
			}

			// Link tag to article
			_, err = tx.ExecContext(
				ctx,
				"INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				article.ID,
				tagID,
			)
			if err != nil {
				return nil, repository.ErrInternal
			}
		}

		// Set the TagList field
		article.TagList = tagList
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	return &article, nil
}

// GetBySlug gets an article by slug
func (r *articleRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*repository.Article, error) {
	var article repository.Article
	article.Author = &repository.User{}
	var authorBio, authorImage sql.NullString

	query := `
		SELECT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at,
		u.id, u.username, u.bio, u.image
		FROM articles a
		JOIN users u ON a.author_id = u.id
		WHERE a.slug = $1
	`

	err := r.db.QueryRowContext(ctx, query, slug).
		Scan(
			&article.ID,
			&article.Slug,
			&article.Title,
			&article.Description,
			&article.Body,
			&article.AuthorID,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.Author.ID,
			&article.Author.Username,
			&authorBio,
			&authorImage,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrArticleNotFound
		}
		return nil, repository.ErrInternal
	}

	// Handle nullable values
	if authorBio.Valid {
		article.Author.Bio = authorBio.String
	}
	if authorImage.Valid {
		article.Author.Image = authorImage.String
	}

	// Get tags for the article
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT t.name FROM tags t JOIN article_tags at ON t.id = at.tag_id WHERE at.article_id = $1",
		article.ID,
	)
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var tagList []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, repository.ErrInternal
		}
		tagList = append(tagList, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.ErrInternal
	}

	article.TagList = tagList

	return &article, nil
}

// Update updates an article
func (r *articleRepository) Update(
	ctx context.Context,
	userID int64,
	slug string,
	title, description, body *string,
) (*repository.Article, error) {
	query := `
		WITH updated_article AS (
			UPDATE articles
			SET
				title = COALESCE($1, title),
				description = COALESCE($2, description),
				body = COALESCE($3, body),
				updated_at = $4
			WHERE slug = $5
			RETURNING id, slug, title, description, body, author_id, created_at, updated_at
		)
		SELECT
			a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at,
			u.id, u.username, u.bio, u.image
		FROM updated_article a
		JOIN users u ON u.id = a.author_id
	`

	now := time.Now()
	var article repository.Article
	article.Author = &repository.User{}
	var authorBio, authorImage sql.NullString

	err := r.db.QueryRowContext(ctx, query, title, description, body, now, slug).
		Scan(
			&article.ID,
			&article.Slug,
			&article.Title,
			&article.Description,
			&article.Body,
			&article.AuthorID,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.Author.ID,
			&article.Author.Username,
			&authorBio,
			&authorImage,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrArticleNotFound
		}
		return nil, repository.ErrInternal
	}

	// Handle nullable values
	if authorBio.Valid {
		article.Author.Bio = authorBio.String
	}
	if authorImage.Valid {
		article.Author.Image = authorImage.String
	}

	// Get tags for the article
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT t.name FROM tags t JOIN article_tags at ON t.id = at.tag_id WHERE at.article_id = $1",
		article.ID,
	)
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var tagList []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, repository.ErrInternal
		}
		tagList = append(tagList, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.ErrInternal
	}

	article.TagList = tagList

	return &article, nil
}

// Delete deletes an article
func (r *articleRepository) Delete(
	ctx context.Context,
	articleID int64,
) error {
	query := `
		DELETE FROM articles
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, articleID)
	if err != nil {
		return repository.ErrInternal
	}

	return nil
}

// Favorite adds an article to the user's favorites
func (r *articleRepository) Favorite(
	ctx context.Context,
	userID int64,
	articleID int64,
) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("transaction rollback error: %v", err)
		}
	}()

	query := `
		INSERT INTO favorites (user_id, article_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err = tx.ExecContext(ctx, query, userID, articleID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" {
				if pqErr.Constraint == "favorites_article_id_fkey" {
					return repository.ErrArticleNotFound
				}
				if pqErr.Constraint == "favorites_user_id_fkey" {
					return repository.ErrUserNotFound
				}
			}
		}
		return repository.ErrInternal
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return repository.ErrInternal
	}

	return nil
}

// Unfavorite removes an article from the user's favorites
func (r *articleRepository) Unfavorite(
	ctx context.Context,
	userID int64,
	articleID int64,
) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("transaction rollback error: %v", err)
		}
	}()

	query := `
		DELETE FROM favorites
		WHERE user_id = $1 AND article_id = $2
	`

	_, err = tx.ExecContext(ctx, query, userID, articleID)
	if err != nil {
		return repository.ErrInternal
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return repository.ErrInternal
	}

	return nil
}

// GetFavoritesCount gets the number of favorites for an article
func (r *articleRepository) GetFavoritesCount(
	ctx context.Context,
	articleID int64,
) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM favorites WHERE article_id = $1", articleID).
		Scan(&count)
	if err != nil {
		return 0, repository.ErrInternal
	}
	return count, nil
}

// IsFavorited checks if a user has favorited an article
func (r *articleRepository) IsFavorited(
	ctx context.Context,
	userID int64,
	articleID int64,
) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM favorites
			WHERE user_id = $1 AND article_id = $2
		)
	`

	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(&exists)
	if err != nil {
		return false, repository.ErrInternal
	}

	return exists, nil
}
