package service

import "conduit/internal/repository"

// MockArticleRepository is a mock implementation of the ArticleRepository interface
type MockArticleRepository struct {
	createArticleFunc func(userID int64, slug, title, description, body string, tagList []string) (*repository.Article, error)
	getBySlugFunc     func(slug string) (*repository.Article, error)
}

// Create is a mock implementation of the Create method
func (m *MockArticleRepository) Create(
	userID int64,
	slug, title, description, body string,
	tagList []string,
) (*repository.Article, error) {
	return m.createArticleFunc(userID, slug, title, description, body, tagList)
}

// GetBySlug is a mock implementation of the GetBySlug method
func (m *MockArticleRepository) GetBySlug(slug string) (*repository.Article, error) {
	return m.getBySlugFunc(slug)
}
