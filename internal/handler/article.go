package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"conduit/internal/middleware"
	"conduit/internal/service"

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
			h.respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Parse request body
		var req CreateArticleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Invalid request body"})
			return
		}

		// Validate request body
		if err := h.validate.Struct(req); err != nil {
			errors := h.translateValidationErrors(err)
			h.respondWithError(w, http.StatusUnprocessableEntity, errors)
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
				h.respondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrArticleAlreadyExists):
				h.respondWithError(
					w,
					http.StatusUnprocessableEntity,
					[]string{"Article with this title already exists"},
				)
			default:
				h.respondWithError(
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
			h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
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
			h.respondWithError(w, http.StatusNotFound, []string{"Article not found"})
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
				h.respondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				h.respondWithError(
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
			h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
		}
	}
}

func (h *articleHandler) translateValidationErrors(err error) []string {
	var validationErrors []string

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			switch e.Tag() {
			case "required":
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("%s is required", e.Field()),
				)
			default:
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("%s is not valid", e.Field()),
				)
			}
		}
	} else {
		validationErrors = append(validationErrors, "Invalid request body")
	}

	return validationErrors
}

// respondWithError sends an error response with the given status code and errors
func (h *articleHandler) respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
