package service

import "conduit/internal/repository"

type MockArticleRepository struct {
	createArticleFunc func(userID int64, title, description, body string, tagList []string) (*repository.Article, error)
}

func (m *MockArticleRepository) Create(userID int64, title, description, body string, tagList []string) (*repository.Article, error) {
	return m.createArticleFunc(userID, title, description, body, tagList)
}
