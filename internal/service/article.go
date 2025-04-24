package service

import (
	"context"
	"errors"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"

	"github.com/gosimple/slug"
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

// ArticleRepository is an interface for the article repository
type ArticleRepository interface {
	Create(
		ctx context.Context,
		userID int64,
		slug, title, description, body string,
		tagList []string,
	) (*repository.Article, error)
	GetBySlug(
		ctx context.Context,
		slug string,
	) (*repository.Article, error)
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

// CreateArticle creates a new article
func (s *articleService) CreateArticle(
	ctx context.Context,
	userID int64,
	title, description, body string,
	tagList []string,
) (*Article, error) {
	// Generate slug from title
	slug := generateSlug(title)

	article, err := s.articleRepository.Create(
		ctx,
		userID,
		slug,
		title,
		description,
		body,
		tagList,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repository.ErrDuplicateSlug):
			return nil, ErrArticleAlreadyExists
		default:
			return nil, ErrInternalServer
		}
	}

	return &Article{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Favorited:      false,
		FavoritesCount: 0,
		Author: Profile{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: false,
		},
	}, nil
}

// GetArticle gets an article by slug
func (s *articleService) GetArticle(
	ctx context.Context,
	slug string,
) (*Article, error) {
	article, err := s.articleRepository.GetBySlug(
		ctx,
		slug,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	return &Article{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Favorited:      false,
		FavoritesCount: 0,
		Author: Profile{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: false,
		},
	}, nil
}

// generateSlug generates a slug from a title
func generateSlug(title string) string {
	return slug.Make(title)
}
