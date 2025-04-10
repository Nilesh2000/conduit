package postgres

import (
	"database/sql"
	"time"

	"conduit/internal/repository"

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
	userID int64,
	slug, title, description, body string,
	tagList []string,
) (*repository.Article, error) {
	// Begin a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return nil, repository.ErrInternal
	}
	defer tx.Rollback()

	query := `
		WITH inserted_article AS (
			INSERT INTO articles (slug, title, description, body, author_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, slug, title, description, body, author_id, created_at, updated_at
		)
		SELECT
			a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at,
			u.username, u.email, u.password, u.bio, u.image
		FROM inserted_article a
		JOIN users u ON u.id = a.author_id
	`

	now := time.Now()
	var article repository.Article
	var authorBio, authorImage sql.NullString

	err = tx.QueryRow(query, slug, title, description, body, userID, now, now).
		Scan(&article.ID, &article.Slug, &article.Title, &article.Description, &article.Body, &article.Author, &article.CreatedAt, &article.UpdatedAt,
			&article.Author.Username, &article.Author.Email, &article.Author.Password, &authorBio, &authorImage)
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
	if len(article.TagList) > 0 {
		for _, tag := range article.TagList {
			var tagID int64
			// Insert or update tag and get its ID
			err := tx.QueryRow("INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id", tag).
				Scan(&tagID)
			if err != nil {
				return nil, repository.ErrInternal
			}

			// Link tag to article
			_, err = tx.Exec(
				"INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				article.ID,
				tagID,
			)
			if err != nil {
				return nil, repository.ErrInternal
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, repository.ErrInternal
	}

	return &article, nil
}
