package postgres

import (
	"conduit/internal/repository"
	"database/sql"
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
func (r *articleRepository) Create(userID int64, title, description, body string, tagList []string) (*repository.Article, error) {
	return nil, nil
}
