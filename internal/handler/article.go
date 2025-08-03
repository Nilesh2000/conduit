package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/repository"
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

// UpdateArticleRequest is the request body for updating an article
type UpdateArticleRequest struct {
	Article struct {
		Title       *string `json:"title" validate:"omitempty"`
		Description *string `json:"description" validate:"omitempty"`
		Body        *string `json:"body" validate:"omitempty"`
	} `json:"article" validate:"required"`
}

// ArticleResponse is the response body for an article
type ArticleResponse struct {
	Article service.Article `json:"article"`
}

// MultipleArticlesResponse is the response body for multiple articles
type MultipleArticlesResponse struct {
	Articles      []service.Article `json:"articles"`
	ArticlesCount int               `json:"articlesCount"`
}

// ArticleService defines the interface for article service operations
type ArticleService interface {
	CreateArticle(
		ctx context.Context,
		userID int64,
		title, description, body string,
		tagList []string,
	) (*service.Article, error)
	GetArticle(ctx context.Context, slug string, currentUserID *int64) (*service.Article, error)
	UpdateArticle(
		ctx context.Context,
		userID int64,
		slug string,
		title, description, body *string,
	) (*service.Article, error)
	DeleteArticle(ctx context.Context, userID int64, slug string) error
	FavoriteArticle(ctx context.Context, userID int64, slug string) (*service.Article, error)
	UnfavoriteArticle(ctx context.Context, userID int64, slug string) (*service.Article, error)
	ListArticles(
		ctx context.Context,
		filters repository.ArticleFilters,
		currentUserID *int64,
	) (*repository.ArticleListResult, error)
	GetArticlesFeed(
		ctx context.Context,
		userID int64,
		limit, offset int,
	) (*repository.ArticleListResult, error)
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

		// Get current user ID from context (optional)
		var userID *int64
		if id, ok := middleware.GetUserIDFromContext(r.Context()); ok {
			userID = &id
		}

		// Get slug from request path
		slug := r.PathValue("slug")

		// Call service to get article
		article, err := h.articleService.GetArticle(
			r.Context(),
			slug,
			userID,
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

func (h *articleHandler) UpdateArticle() http.HandlerFunc {
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
		var req UpdateArticleRequest
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

		// Get slug from request path
		slug := r.PathValue("slug")

		// Call service to update article
		article, err := h.articleService.UpdateArticle(
			r.Context(),
			userID,
			slug,
			req.Article.Title,
			req.Article.Description,
			req.Article.Body,
		)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrArticleNotAuthorized):
				response.RespondWithError(
					w,
					http.StatusForbidden,
					[]string{"You are not the author of this article"},
				)
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

		// Respond with updated article
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

// DeleteArticle is a handler function for deleting an article
func (h *articleHandler) DeleteArticle() http.HandlerFunc {
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

		// Call service to delete article
		err := h.articleService.DeleteArticle(r.Context(), userID, slug)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrArticleNotAuthorized):
				response.RespondWithError(
					w,
					http.StatusForbidden,
					[]string{"You are not the author of this article"},
				)
			case errors.Is(err, service.ErrArticleNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
		}

		// Respond with deleted article
		w.WriteHeader(http.StatusOK)
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

// ListArticles is a handler function for listing articles with optional filters
func (h *articleHandler) ListArticles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get current user ID from context (optional)
		var userID *int64
		if id, ok := middleware.GetUserIDFromContext(r.Context()); ok {
			userID = &id
		}

		// Parse query parameters
		filters := repository.ArticleFilters{
			Limit:  20, // Default limit
			Offset: 0,  // Default offset
		}

		// Parse tag filter
		if tag := r.URL.Query().Get("tag"); tag != "" {
			filters.Tag = &tag
		}

		// Parse author filter
		if author := r.URL.Query().Get("author"); author != "" {
			filters.Author = &author
		}

		// Parse favorited filter
		if favorited := r.URL.Query().Get("favorited"); favorited != "" {
			filters.Favorited = &favorited
		}

		// Parse limit parameter
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				filters.Limit = limit
			}
		}

		// Parse offset parameter
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				filters.Offset = offset
			}
		}

		// Call service to list articles
		result, err := h.articleService.ListArticles(r.Context(), filters, userID)
		if err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
			return
		}

		// Convert repository articles to service articles
		var articles []service.Article
		for _, repoArticle := range result.Articles {
			article := service.Article{
				Slug:           repoArticle.Slug,
				Title:          repoArticle.Title,
				Description:    repoArticle.Description,
				Body:           repoArticle.Body,
				TagList:        repoArticle.TagList,
				CreatedAt:      repoArticle.CreatedAt,
				UpdatedAt:      repoArticle.UpdatedAt,
				Favorited:      repoArticle.Favorited,
				FavoritesCount: repoArticle.FavoritesCount,
				Author: service.Profile{
					Username:  repoArticle.Author.Username,
					Bio:       repoArticle.Author.Bio,
					Image:     repoArticle.Author.Image,
					Following: repoArticle.Author.Following,
				},
			}
			articles = append(articles, article)
		}

		// Respond with articles
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(MultipleArticlesResponse{
			Articles:      articles,
			ArticlesCount: result.Count,
		}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// GetArticlesFeed is a handler function for getting articles from followed users
func (h *articleHandler) GetArticlesFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context (required for feed)
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Parse query parameters
		limit := 20 // Default limit
		offset := 0 // Default offset

		// Parse limit parameter
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		// Parse offset parameter
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		// Call service to get articles feed
		result, err := h.articleService.GetArticlesFeed(r.Context(), userID, limit, offset)
		if err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
			return
		}

		// Convert repository articles to service articles
		var articles []service.Article
		for _, repoArticle := range result.Articles {
			article := service.Article{
				Slug:           repoArticle.Slug,
				Title:          repoArticle.Title,
				Description:    repoArticle.Description,
				Body:           repoArticle.Body,
				TagList:        repoArticle.TagList,
				CreatedAt:      repoArticle.CreatedAt,
				UpdatedAt:      repoArticle.UpdatedAt,
				Favorited:      repoArticle.Favorited,
				FavoritesCount: repoArticle.FavoritesCount,
				Author: service.Profile{
					Username:  repoArticle.Author.Username,
					Bio:       repoArticle.Author.Bio,
					Image:     repoArticle.Author.Image,
					Following: repoArticle.Author.Following,
				},
			}
			articles = append(articles, article)
		}

		// Respond with articles
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(MultipleArticlesResponse{
			Articles:      articles,
			ArticlesCount: result.Count,
		}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}
