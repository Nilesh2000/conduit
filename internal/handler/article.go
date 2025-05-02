package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
	"github.com/Nilesh2000/conduit/internal/validation"

	"github.com/go-playground/validator/v10"
)

// CreateArticleRequest is the request body for creating an article
type CreateArticleRequest struct {
	Article struct {
		Title       string   `json:"title" validate:"required"`
		Description string   `json:"description" validate:"required"`
		Body        string   `json:"body" validate:"required"`
		TagList     []string `json:"tagList,omitempty"`
	} `json:"article" validate:"required"`
}

// ArticleResponse is the response body for an article
type ArticleResponse struct {
	Article service.Article `json:"article"`
}

// ArticleService defines the interface for article service operations
type ArticleService interface {
	CreateArticle(
		ctx context.Context,
		userID int64,
		title, description, body string,
		tagList []string,
	) (*service.Article, error)
	GetArticle(ctx context.Context, slug string) (*service.Article, error)
	FavoriteArticle(ctx context.Context, userID int64, slug string) (*service.Article, error)
	UnfavoriteArticle(ctx context.Context, userID int64, slug string) (*service.Article, error)
}

// articleHandler is a handler for article operations
type articleHandler struct {
	articleService ArticleService
	validate       *validator.Validate
}

// NewArticleHandler creates a new ArticleHandler
func NewArticleHandler(articleService ArticleService) *articleHandler {
	return &articleHandler{
		articleService: articleService,
		validate:       validator.New(),
	}
}

// CreateArticle is a handler function for creating a new article
func (h *articleHandler) CreateArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Parse request body
		var req CreateArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.RespondWithError(
				w,
				http.StatusUnprocessableEntity,
				[]string{"Invalid request body"},
			)
			return
		}

		// Validate request body
		if err := h.validate.Struct(req); err != nil {
			errors := validation.TranslateValidationErrors(err)
			response.RespondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to create article
		article, err := h.articleService.CreateArticle(
			r.Context(),
			userID,
			req.Article.Title,
			req.Article.Description,
			req.Article.Body,
			req.Article.TagList,
		)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrArticleAlreadyExists):
				response.RespondWithError(
					w,
					http.StatusUnprocessableEntity,
					[]string{"Article with this title already exists"},
				)
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with created article
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(ArticleResponse{Article: *article}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// GetArticle is a handler function for getting an article by slug
func (h *articleHandler) GetArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get slug from request path
		slug := r.URL.Path[len("/api/articles/"):]
		if slug == "" {
			response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			return
		}

		// Call service to get article
		article, err := h.articleService.GetArticle(
			r.Context(),
			slug,
		)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrArticleNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with article
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ArticleResponse{Article: *article}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// FavoriteArticle is a handler function for favoriting an article
func (h *articleHandler) FavoriteArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get slug from request path
		slug := r.PathValue("slug")

		// Call service to favorite article
		article, err := h.articleService.FavoriteArticle(
			r.Context(),
			userID,
			slug,
		)
		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrArticleNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with favorite article
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ArticleResponse{Article: *article}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// UnfavoriteArticle is a handler function for unfavoriting an article
func (h *articleHandler) UnfavoriteArticle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get slug from request path
		slug := r.PathValue("slug")

		// Call service to unfavorite article
		article, err := h.articleService.UnfavoriteArticle(
			r.Context(),
			userID,
			slug,
		)
		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrArticleNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with unfavorite article
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ArticleResponse{Article: *article}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}
