package handler

import "conduit/internal/service"

// ArticleService defines the interface for article service operations
type ArticleService struct {
	CreateArticle func(userID int64, title, description, body string, tagList []string) (*service.Article, error)
}
