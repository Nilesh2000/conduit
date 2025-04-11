package service

import (
	"conduit/internal/repository"
	"time"
)

// Article represents a article
type Article struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int       `json:"favoritesCount"`
	Author         Profile   `json:"author"`
}

// Profile represents a user profile
type Profile struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

type ArticleRepository interface {
	Create(userID int64, title, description, body string, tagList []string) (*repository.Article, error)
}

// articleService implements the articleService interface
type articleService struct {
	articleRepository ArticleRepository
}

// NewArticleService creates a new ArticleService
func NewArticleService(articleRepository ArticleRepository) *articleService {
	return &articleService{
		articleRepository: articleRepository,
	}
}

func (s *articleService) CreateArticle(userID int64, title, description, body string, tagList []string) (*Article, error) {
	return nil, nil
}
