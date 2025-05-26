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
	Update(
		ctx context.Context,
		userID int64,
		slug string,
		title, description, body *string,
	) (*repository.Article, error)
	Delete(
		ctx context.Context,
		articleID int64,
	) error
	Favorite(
		ctx context.Context,
		userID int64,
		articleID int64,
	) error
	Unfavorite(
		ctx context.Context,
		userID int64,
		articleID int64,
	) error
	GetFavoritesCount(
		ctx context.Context,
		articleID int64,
	) (int, error)
	IsFavorited(
		ctx context.Context,
		userID int64,
		articleID int64,
	) (bool, error)
}

// articleService implements the articleService interface
type articleService struct {
	articleRepository ArticleRepository
	profileRepository ProfileRepository
}

// NewArticleService creates a new ArticleService
func NewArticleService(
	articleRepository ArticleRepository,
	profileRepository ProfileRepository,
) *articleService {
	return &articleService{
		articleRepository: articleRepository,
		profileRepository: profileRepository,
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
	currentUserID *int64,
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

	// Get favorites count
	favoritesCount, err := s.articleRepository.GetFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Check if user has favorited the article
	favorited := false
	if currentUserID != nil {
		favorited, err = s.articleRepository.IsFavorited(ctx, *currentUserID, article.ID)
		if err != nil {
			return nil, ErrInternalServer
		}
	}

	// Check if user is following the author
	following := false
	if currentUserID != nil {
		following, err = s.profileRepository.IsFollowing(ctx, *currentUserID, article.Author.ID)
		if err != nil {
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
		Favorited:      favorited,
		FavoritesCount: favoritesCount,
		Author: Profile{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: following,
		},
	}, nil
}

// UpdateArticle updates an article
func (s *articleService) UpdateArticle(
	ctx context.Context,
	userID int64,
	slug string,
	title, description, body *string,
) (*Article, error) {
	article, err := s.articleRepository.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	// Check if user is the author
	if article.AuthorID != userID {
		return nil, ErrArticleNotAuthorized
	}

	// Proceed with update
	article, err = s.articleRepository.Update(
		ctx,
		userID,
		slug,
		title,
		description,
		body,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	// Get favorites count
	favoritesCount, err := s.articleRepository.GetFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, ErrInternalServer
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
		FavoritesCount: favoritesCount,
		Author: Profile{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: false,
		},
	}, nil
}

// DeleteArticle deletes an article
func (s *articleService) DeleteArticle(
	ctx context.Context,
	userID int64,
	slug string,
) error {
	article, err := s.articleRepository.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return ErrArticleNotFound
		default:
			return ErrInternalServer
		}
	}

	// Check if user is the author
	if article.AuthorID != userID {
		return ErrArticleNotAuthorized
	}

	err = s.articleRepository.Delete(ctx, article.ID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return ErrArticleNotFound
		default:
			return ErrInternalServer
		}
	}

	return nil
}

// FavoriteArticle favorites an article
func (s *articleService) FavoriteArticle(
	ctx context.Context,
	userID int64,
	slug string,
) (*Article, error) {
	article, err := s.articleRepository.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	// Check if author is the same as the user
	if article.Author.ID == userID {
		return nil, ErrArticleAuthorCannotFavorite
	}

	// Favorite the article
	err = s.articleRepository.Favorite(ctx, userID, article.ID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	// Get favorites count
	favoritesCount, err := s.articleRepository.GetFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Check if user is following the author
	following, err := s.profileRepository.IsFollowing(ctx, userID, article.Author.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	return &Article{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Favorited:      true,
		FavoritesCount: favoritesCount,
		Author: Profile{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: following,
		},
	}, nil
}

// UnfavoriteArticle unfavorites an article
func (s *articleService) UnfavoriteArticle(
	ctx context.Context,
	userID int64,
	slug string,
) (*Article, error) {
	article, err := s.articleRepository.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	// Unfavorite the article
	err = s.articleRepository.Unfavorite(ctx, userID, article.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Get favorites count
	favoritesCount, err := s.articleRepository.GetFavoritesCount(ctx, article.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Check if user is following the author
	following, err := s.profileRepository.IsFollowing(ctx, userID, article.Author.ID)
	if err != nil {
		return nil, ErrInternalServer
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
		FavoritesCount: favoritesCount,
		Author: Profile{
			Username:  article.Author.Username,
			Bio:       article.Author.Bio,
			Image:     article.Author.Image,
			Following: following,
		},
	}, nil
}

// generateSlug generates a slug from a title
func generateSlug(title string) string {
	return slug.Make(title)
}
