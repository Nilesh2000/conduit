package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/repository"
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
)

// MockArticleService is a mock implementation of the ArticleService interface
type MockArticleService struct {
	createArticleFunc     func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error)
	getArticleFunc        func(ctx context.Context, slug string, currentUserID *int64) (*service.Article, error)
	updateArticleFunc     func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error)
	deleteArticleFunc     func(ctx context.Context, userID int64, slug string) error
	favoriteArticleFunc   func(ctx context.Context, userID int64, slug string) (*service.Article, error)
	unfavoriteArticleFunc func(ctx context.Context, userID int64, slug string) (*service.Article, error)
	listArticlesFunc      func(ctx context.Context, filters repository.ArticleFilters, currentUserID *int64) (*repository.ArticleListResult, error)
	getArticlesFeedFunc   func(ctx context.Context, userID int64, limit, offset int) (*repository.ArticleListResult, error)
}

// CreateArticle is a mock implementation of the CreateArticle method
func (m *MockArticleService) CreateArticle(
	ctx context.Context,
	userID int64,
	title, description, body string,
	tagList []string,
) (*service.Article, error) {
	return m.createArticleFunc(ctx, userID, title, description, body, tagList)
}

// GetArticle is a mock implementation of the GetArticle method
func (m *MockArticleService) GetArticle(
	ctx context.Context,
	slug string,
	currentUserID *int64,
) (*service.Article, error) {
	return m.getArticleFunc(ctx, slug, currentUserID)
}

// UpdateArticle is a mock implementation of the UpdateArticle method
func (m *MockArticleService) UpdateArticle(
	ctx context.Context,
	userID int64,
	slug string,
	title, description, body *string,
) (*service.Article, error) {
	return m.updateArticleFunc(ctx, userID, slug, title, description, body)
}

// DeleteArticle is a mock implementation of the DeleteArticle method
func (m *MockArticleService) DeleteArticle(
	ctx context.Context,
	userID int64,
	slug string,
) error {
	return m.deleteArticleFunc(ctx, userID, slug)
}

// FavoriteArticle is a mock implementation of the FavoriteArticle method
func (m *MockArticleService) FavoriteArticle(
	ctx context.Context,
	userID int64,
	slug string,
) (*service.Article, error) {
	return m.favoriteArticleFunc(ctx, userID, slug)
}

// UnfavoriteArticle is a mock implementation of the UnfavoriteArticle method
func (m *MockArticleService) UnfavoriteArticle(
	ctx context.Context,
	userID int64,
	slug string,
) (*service.Article, error) {
	return m.unfavoriteArticleFunc(ctx, userID, slug)
}

// ListArticles is a mock implementation of the ListArticles method
func (m *MockArticleService) ListArticles(
	ctx context.Context,
	filters repository.ArticleFilters,
	currentUserID *int64,
) (*repository.ArticleListResult, error) {
	return m.listArticlesFunc(ctx, filters, currentUserID)
}

// GetArticlesFeed is a mock implementation of the GetArticlesFeed method
func (m *MockArticleService) GetArticlesFeed(
	ctx context.Context,
	userID int64,
	limit, offset int,
) (*repository.ArticleListResult, error) {
	return m.getArticlesFeedFunc(ctx, userID, limit, offset)
}

// TestArticleHandler_CreateArticle tests the CreateArticle method of the ArticleHandler
func TestArticleHandler_CreateArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestBody      any
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successful article creation",
			requestBody: CreateArticleRequest{
				Article: struct {
					Title       string   `json:"title" validate:"required"`
					Description string   `json:"description" validate:"required"`
					Body        string   `json:"body" validate:"required"`
					TagList     []string `json:"tagList,omitempty"`
				}{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if title != "Test Article" {
							t.Errorf("Expected title 'Test Article', got %s", title)
						}
						if description != "Test Description" {
							t.Errorf("Expected description 'Test Description', got %s", description)
						}
						if body != "Test Body" {
							t.Errorf("Expected body 'Test Body', got %s", body)
						}
						if len(tagList) != 2 || tagList[0] != "tag1" || tagList[1] != "tag2" {
							t.Errorf("Expected tagList [tag1, tag2], got %v", tagList)
						}

						now := time.Now()
						return &service.Article{
							Slug:           "test-article",
							Title:          title,
							Description:    description,
							Body:           body,
							TagList:        tagList,
							CreatedAt:      now,
							UpdatedAt:      now,
							Favorited:      false,
							FavoritesCount: 0,
							Author: service.Profile{
								Username:  "testuser",
								Bio:       "Test Bio",
								Image:     "https://example.com/image.jpg",
								Following: false,
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: ArticleResponse{
				Article: service.Article{
					Slug:           "test-article",
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					CreatedAt:      time.Time{},
					UpdatedAt:      time.Time{},
					Favorited:      false,
					FavoritesCount: 0,
					Author: service.Profile{
						Username:  "testuser",
						Bio:       "Test Bio",
						Image:     "https://example.com/image.jpg",
						Following: false,
					},
				},
			},
		},
		{
			name: "Unauthenticated request",
			requestBody: CreateArticleRequest{
				Article: struct {
					Title       string   `json:"title" validate:"required"`
					Description string   `json:"description" validate:"required"`
					Body        string   `json:"body" validate:"required"`
					TagList     []string `json:"tagList,omitempty"`
				}{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						t.Errorf("CreateArticle should not be called for unauthenticated request")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name: "Invalid JSON",
			requestBody: `{
				"article": {
					"title": "Test Article",
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						t.Errorf("CreateArticle should not be called for invalid JSON")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Invalid request body"}},
			},
		},
		{
			name: "Missing required fields",
			requestBody: CreateArticleRequest{
				Article: struct {
					Title       string   `json:"title" validate:"required"`
					Description string   `json:"description" validate:"required"`
					Body        string   `json:"body" validate:"required"`
					TagList     []string `json:"tagList,omitempty"`
				}{
					Title: "Test Article",
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						t.Errorf("CreateArticle should not be called for missing required fields")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Description is required", "Body is required"}},
			},
		},
		{
			name: "Article already exists",
			requestBody: CreateArticleRequest{
				Article: struct {
					Title       string   `json:"title" validate:"required"`
					Description string   `json:"description" validate:"required"`
					Body        string   `json:"body" validate:"required"`
					TagList     []string `json:"tagList,omitempty"`
				}{
					Title:       "Existing Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						return nil, service.ErrArticleAlreadyExists
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Article with this title already exists"}},
			},
		},
		{
			name: "User not found",
			requestBody: CreateArticleRequest{
				Article: struct {
					Title       string   `json:"title" validate:"required"`
					Description string   `json:"description" validate:"required"`
					Body        string   `json:"body" validate:"required"`
					TagList     []string `json:"tagList,omitempty"`
				}{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						return nil, service.ErrUserNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"User not found"}},
			},
		},
		{
			name: "Internal server error",
			requestBody: CreateArticleRequest{
				Article: struct {
					Title       string   `json:"title" validate:"required"`
					Description string   `json:"description" validate:"required"`
					Body        string   `json:"body" validate:"required"`
					TagList     []string `json:"tagList,omitempty"`
				}{
					Title:       "Test Article",
					Description: "Test Description",
					Body:        "Test Body",
					TagList:     []string{"tag1", "tag2"},
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					createArticleFunc: func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			var bodyBytes []byte

			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create Request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/articles",
				bytes.NewReader(bodyBytes),
			)
			req.Header.Set("Content-Type", "application/json")

			// Setup Auth if provided
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.CreateArticle()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusCreated {
				var resp ArticleResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				resp.Article.CreatedAt = time.Time{}
				resp.Article.UpdatedAt = time.Time{}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}

// TestArticleHandler_GetArticle tests the GetArticle method of the ArticleHandler
func TestArticleHandler_GetArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Article found",
			slug: "test-article",
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticleFunc: func(ctx context.Context, slug string, currentUserID *int64) (*service.Article, error) {
						if slug != "test-article" {
							t.Errorf("Expected slug 'test-article', got %q", slug)
						}

						now := time.Now()
						return &service.Article{
							Slug:           "test-article",
							Title:          "Test Article",
							Description:    "Test Description",
							Body:           "Test Body",
							TagList:        []string{"tag1", "tag2"},
							CreatedAt:      now,
							UpdatedAt:      now,
							Favorited:      false,
							FavoritesCount: 0,
							Author: service.Profile{
								Username:  "testuser",
								Bio:       "Test Bio",
								Image:     "https://example.com/image.jpg",
								Following: false,
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ArticleResponse{
				Article: service.Article{
					Slug:           "test-article",
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					CreatedAt:      time.Time{},
					UpdatedAt:      time.Time{},
					Favorited:      false,
					FavoritesCount: 0,
					Author: service.Profile{
						Username:  "testuser",
						Bio:       "Test Bio",
						Image:     "https://example.com/image.jpg",
						Following: false,
					},
				},
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticleFunc: func(ctx context.Context, slug string, currentUserID *int64) (*service.Article, error) {
						return nil, service.ErrArticleNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Article not found"}},
			},
		},
		{
			name: "Internal server error",
			slug: "test-article",
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticleFunc: func(ctx context.Context, slug string, currentUserID *int64) (*service.Article, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/articles/"+tt.slug,
				nil,
			)
			req.SetPathValue("slug", tt.slug)

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.GetArticle()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp ArticleResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				resp.Article.CreatedAt = time.Time{}
				resp.Article.UpdatedAt = time.Time{}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}

// TestArticleHandler_UpdateArticle tests the UpdateArticle method of the ArticleHandler
func TestArticleHandler_UpdateArticle(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string {
		return &s
	}

	tests := []struct {
		name             string
		slug             string
		requestBody      any
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successfully update an article",
			slug: "test-article",
			requestBody: UpdateArticleRequest{
				Article: struct {
					Title       *string `json:"title" validate:"omitempty"`
					Description *string `json:"description" validate:"omitempty"`
					Body        *string `json:"body" validate:"omitempty"`
				}{
					Title:       strPtr("Updated Title"),
					Description: strPtr("Updated Description"),
					Body:        strPtr("Updated Body"),
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					updateArticleFunc: func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if slug != "test-article" {
							t.Errorf("Expected slug 'test-article', got %q", slug)
						}
						if *title != "Updated Title" {
							t.Errorf("Expected title 'Updated Title', got %q", *title)
						}
						if *description != "Updated Description" {
							t.Errorf(
								"Expected description 'Updated Description', got %q",
								*description,
							)
						}
						if *body != "Updated Body" {
							t.Errorf("Expected body 'Updated Body', got %q", *body)
						}

						now := time.Now()
						return &service.Article{
							Slug:           "test-article",
							Title:          "Updated Title",
							Description:    "Updated Description",
							Body:           "Updated Body",
							TagList:        []string{"tag1", "tag2"},
							CreatedAt:      now,
							UpdatedAt:      now,
							Favorited:      false,
							FavoritesCount: 0,
							Author: service.Profile{
								Username:  "testuser",
								Bio:       "Test Bio",
								Image:     "https://example.com/image.jpg",
								Following: false,
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ArticleResponse{
				Article: service.Article{
					Slug:           "test-article",
					Title:          "Updated Title",
					Description:    "Updated Description",
					Body:           "Updated Body",
					TagList:        []string{"tag1", "tag2"},
					CreatedAt:      time.Time{},
					UpdatedAt:      time.Time{},
					Favorited:      false,
					FavoritesCount: 0,
					Author: service.Profile{
						Username:  "testuser",
						Bio:       "Test Bio",
						Image:     "https://example.com/image.jpg",
						Following: false,
					},
				},
			},
		},
		{
			name: "Unauthenticated request",
			slug: "test-article",
			requestBody: UpdateArticleRequest{
				Article: struct {
					Title       *string `json:"title" validate:"omitempty"`
					Description *string `json:"description" validate:"omitempty"`
					Body        *string `json:"body" validate:"omitempty"`
				}{
					Title:       strPtr("Updated Title"),
					Description: strPtr("Updated Description"),
					Body:        strPtr("Updated Body"),
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					updateArticleFunc: func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error) {
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name: "Invalid request body",
			slug: "test-article",
			requestBody: `{
				"article": {
					"title": "Updated Title",
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					updateArticleFunc: func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error) {
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Invalid request body"}},
			},
		},
		{
			name: "Not the author of the article",
			slug: "test-article",
			requestBody: UpdateArticleRequest{
				Article: struct {
					Title       *string `json:"title" validate:"omitempty"`
					Description *string `json:"description" validate:"omitempty"`
					Body        *string `json:"body" validate:"omitempty"`
				}{
					Title:       strPtr("Updated Title"),
					Description: strPtr("Updated Description"),
					Body:        strPtr("Updated Body"),
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(2))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					updateArticleFunc: func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error) {
						return nil, service.ErrArticleNotAuthorized
					},
				}
				return mockService
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"You are not the author of this article"}},
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			requestBody: UpdateArticleRequest{
				Article: struct {
					Title       *string `json:"title" validate:"omitempty"`
					Description *string `json:"description" validate:"omitempty"`
					Body        *string `json:"body" validate:"omitempty"`
				}{
					Title:       strPtr("Updated Title"),
					Description: strPtr("Updated Description"),
					Body:        strPtr("Updated Body"),
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					updateArticleFunc: func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error) {
						return nil, service.ErrArticleNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Article not found"}},
			},
		},
		{
			name: "Internal server error",
			slug: "test-article",
			requestBody: UpdateArticleRequest{
				Article: struct {
					Title       *string `json:"title" validate:"omitempty"`
					Description *string `json:"description" validate:"omitempty"`
					Body        *string `json:"body" validate:"omitempty"`
				}{
					Title:       strPtr("Updated Title"),
					Description: strPtr("Updated Description"),
					Body:        strPtr("Updated Body"),
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					updateArticleFunc: func(ctx context.Context, userID int64, slug string, title, description, body *string) (*service.Article, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock Service
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			var bodyBytes []byte

			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create Request
			req := httptest.NewRequest(
				http.MethodPut,
				"/api/articles/"+tt.slug,
				bytes.NewReader(bodyBytes),
			)
			req.SetPathValue("slug", tt.slug)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.UpdateArticle()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp ArticleResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				resp.Article.CreatedAt = time.Time{}
				resp.Article.UpdatedAt = time.Time{}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}

// TestArticleHandler_DeleteArticle tests the DeleteArticle method of the ArticleHandler
func TestArticleHandler_DeleteArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successfully delete an article",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					deleteArticleFunc: func(ctx context.Context, userID int64, slug string) error {
						return nil
					},
				}
				return mockService
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: nil,
		},
		{
			name: "Unauthenticated request",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					deleteArticleFunc: func(ctx context.Context, userID int64, slug string) error {
						return nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name: "Not the author of the article",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(2))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					deleteArticleFunc: func(ctx context.Context, userID int64, slug string) error {
						return service.ErrArticleNotAuthorized
					},
				}
				return mockService
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"You are not the author of this article"}},
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					deleteArticleFunc: func(ctx context.Context, userID int64, slug string) error {
						return service.ErrArticleNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Article not found"}},
			},
		},
		{
			name: "Internal server error",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					deleteArticleFunc: func(ctx context.Context, userID int64, slug string) error {
						return service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock Service
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodDelete,
				"/api/articles/"+tt.slug,
				nil,
			)
			req.SetPathValue("slug", tt.slug)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.DeleteArticle()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				got = nil
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}

// TestArticleHandler_FavoriteArticle tests the FavoriteArticle method of the ArticleHandler
func TestArticleHandler_FavoriteArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successfully favorite an article",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					favoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if slug != "test-article" {
							t.Errorf("Expected slug 'test-article', got %q", slug)
						}
						now := time.Now()
						return &service.Article{
							Slug:           "test-article",
							Title:          "Test Article",
							Description:    "Test Description",
							Body:           "Test Body",
							TagList:        []string{"tag1", "tag2"},
							CreatedAt:      now,
							UpdatedAt:      now,
							Favorited:      true,
							FavoritesCount: 1,
							Author: service.Profile{
								Username:  "testuser",
								Bio:       "Test Bio",
								Image:     "https://example.com/image.jpg",
								Following: false,
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ArticleResponse{
				Article: service.Article{
					Slug:           "test-article",
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					CreatedAt:      time.Time{},
					UpdatedAt:      time.Time{},
					Favorited:      true,
					FavoritesCount: 1,
					Author: service.Profile{
						Username:  "testuser",
						Bio:       "Test Bio",
						Image:     "https://example.com/image.jpg",
						Following: false,
					},
				},
			},
		},
		{
			name: "Unauthenticated request",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					favoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						t.Errorf("FavoriteArticle should not be called for unauthenticated request")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					favoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						return nil, service.ErrArticleNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Article not found"}},
			},
		},
		{
			name: "Internal server error",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					favoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/articles/"+tt.slug+"/favorite",
				nil,
			)
			req.SetPathValue("slug", tt.slug)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.FavoriteArticle()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp ArticleResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				resp.Article.CreatedAt = time.Time{}
				resp.Article.UpdatedAt = time.Time{}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}

// TestArticleHandler_UnfavoriteArticle tests the UnfavoriteArticle method of the ArticleHandler
func TestArticleHandler_UnfavoriteArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successfully unfavorite an article",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					unfavoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if slug != "test-article" {
							t.Errorf("Expected slug 'test-article', got %q", slug)
						}
						now := time.Now()
						return &service.Article{
							Slug:           "test-article",
							Title:          "Test Article",
							Description:    "Test Description",
							Body:           "Test Body",
							TagList:        []string{"tag1", "tag2"},
							CreatedAt:      now,
							UpdatedAt:      now,
							Favorited:      false,
							FavoritesCount: 0,
							Author: service.Profile{
								Username:  "testuser",
								Bio:       "Test Bio",
								Image:     "https://example.com/image.jpg",
								Following: false,
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ArticleResponse{
				Article: service.Article{
					Slug:           "test-article",
					Title:          "Test Article",
					Description:    "Test Description",
					Body:           "Test Body",
					TagList:        []string{"tag1", "tag2"},
					CreatedAt:      time.Time{},
					UpdatedAt:      time.Time{},
					Favorited:      false,
					FavoritesCount: 0,
					Author: service.Profile{
						Username:  "testuser",
						Bio:       "Test Bio",
						Image:     "https://example.com/image.jpg",
						Following: false,
					},
				},
			},
		},
		{
			name: "Unauthenticated request",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					unfavoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						t.Errorf(
							"UnfavoriteArticle should not be called for unauthenticated request",
						)
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					unfavoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						return nil, service.ErrArticleNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Article not found"}},
			},
		},
		{
			name: "Internal server error",
			slug: "test-article",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					unfavoriteArticleFunc: func(ctx context.Context, userID int64, slug string) (*service.Article, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodDelete,
				"/api/articles/"+tt.slug+"/favorite",
				nil,
			)
			req.SetPathValue("slug", tt.slug)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.UnfavoriteArticle()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp ArticleResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				resp.Article.CreatedAt = time.Time{}
				resp.Article.UpdatedAt = time.Time{}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			// Deep compare expected and got
			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}

// TestArticleHandler_ListArticles tests the ListArticles method of the ArticleHandler
func TestArticleHandler_ListArticles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		queryParams      string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name:        "Successfully list articles without filters",
			queryParams: "",
			setupAuth: func(r *http.Request) *http.Request {
				// Optional authentication
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					listArticlesFunc: func(ctx context.Context, filters repository.ArticleFilters, currentUserID *int64) (*repository.ArticleListResult, error) {
						if filters.Limit != 20 {
							t.Errorf("Expected limit 20, got %d", filters.Limit)
						}
						if filters.Offset != 0 {
							t.Errorf("Expected offset 0, got %d", filters.Offset)
						}

						now := time.Now()
						articles := []*repository.Article{
							{
								ID:             1,
								Slug:           "test-article-1",
								Title:          "Test Article 1",
								Description:    "Test Description 1",
								Body:           "Test Body 1",
								TagList:        []string{"tag1", "tag2"},
								CreatedAt:      now,
								UpdatedAt:      now,
								Favorited:      false,
								FavoritesCount: 0,
								Author: &repository.User{
									ID:        1,
									Username:  "testuser1",
									Bio:       "Test Bio 1",
									Image:     "https://example.com/image1.jpg",
									Following: false,
								},
							},
							{
								ID:             2,
								Slug:           "test-article-2",
								Title:          "Test Article 2",
								Description:    "Test Description 2",
								Body:           "Test Body 2",
								TagList:        []string{"tag3"},
								CreatedAt:      now,
								UpdatedAt:      now,
								Favorited:      true,
								FavoritesCount: 5,
								Author: &repository.User{
									ID:        2,
									Username:  "testuser2",
									Bio:       "Test Bio 2",
									Image:     "https://example.com/image2.jpg",
									Following: true,
								},
							},
						}

						return &repository.ArticleListResult{
							Articles: articles,
							Count:    2,
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: MultipleArticlesResponse{
				Articles: []service.Article{
					{
						Slug:           "test-article-1",
						Title:          "Test Article 1",
						Description:    "Test Description 1",
						Body:           "Test Body 1",
						TagList:        []string{"tag1", "tag2"},
						CreatedAt:      time.Time{},
						UpdatedAt:      time.Time{},
						Favorited:      false,
						FavoritesCount: 0,
						Author: service.Profile{
							Username:  "testuser1",
							Bio:       "Test Bio 1",
							Image:     "https://example.com/image1.jpg",
							Following: false,
						},
					},
					{
						Slug:           "test-article-2",
						Title:          "Test Article 2",
						Description:    "Test Description 2",
						Body:           "Test Body 2",
						TagList:        []string{"tag3"},
						CreatedAt:      time.Time{},
						UpdatedAt:      time.Time{},
						Favorited:      true,
						FavoritesCount: 5,
						Author: service.Profile{
							Username:  "testuser2",
							Bio:       "Test Bio 2",
							Image:     "https://example.com/image2.jpg",
							Following: true,
						},
					},
				},
				ArticlesCount: 2,
			},
		},
		{
			name:        "Successfully list articles with filters",
			queryParams: "?tag=test&author=testuser&limit=10&offset=5",
			setupAuth: func(r *http.Request) *http.Request {
				// Optional authentication
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					listArticlesFunc: func(ctx context.Context, filters repository.ArticleFilters, currentUserID *int64) (*repository.ArticleListResult, error) {
						if *filters.Tag != "test" {
							t.Errorf("Expected tag 'test', got %s", *filters.Tag)
						}
						if *filters.Author != "testuser" {
							t.Errorf("Expected author 'testuser', got %s", *filters.Author)
						}
						if filters.Limit != 10 {
							t.Errorf("Expected limit 10, got %d", filters.Limit)
						}
						if filters.Offset != 5 {
							t.Errorf("Expected offset 5, got %d", filters.Offset)
						}

						return &repository.ArticleListResult{
							Articles: []*repository.Article{},
							Count:    0,
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: MultipleArticlesResponse{
				Articles:      []service.Article{},
				ArticlesCount: 0,
			},
		},
		{
			name:        "Internal server error",
			queryParams: "",
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					listArticlesFunc: func(ctx context.Context, filters repository.ArticleFilters, currentUserID *int64) (*repository.ArticleListResult, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/articles"+tt.queryParams,
				nil,
			)

			// Setup Auth if provided
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.ListArticles()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			if tt.expectedStatus == http.StatusOK {
				var resp MultipleArticlesResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				for i := range resp.Articles {
					resp.Articles[i].CreatedAt = time.Time{}
					resp.Articles[i].UpdatedAt = time.Time{}
				}

				expected := tt.expectedResponse.(MultipleArticlesResponse)
				if resp.ArticlesCount != expected.ArticlesCount {
					t.Errorf(
						"ArticlesCount: got %d, want %d",
						resp.ArticlesCount,
						expected.ArticlesCount,
					)
				}
				if len(resp.Articles) != len(expected.Articles) {
					t.Errorf(
						"Articles length: got %d, want %d",
						len(resp.Articles),
						len(expected.Articles),
					)
				}
				for i, article := range resp.Articles {
					if i >= len(expected.Articles) {
						break
					}
					expectedArticle := expected.Articles[i]
					if article.Slug != expectedArticle.Slug {
						t.Errorf(
							"Article[%d].Slug: got %s, want %s",
							i,
							article.Slug,
							expectedArticle.Slug,
						)
					}
					if article.Title != expectedArticle.Title {
						t.Errorf(
							"Article[%d].Title: got %s, want %s",
							i,
							article.Title,
							expectedArticle.Title,
						)
					}
					if article.Favorited != expectedArticle.Favorited {
						t.Errorf(
							"Article[%d].Favorited: got %t, want %t",
							i,
							article.Favorited,
							expectedArticle.Favorited,
						)
					}
					if article.FavoritesCount != expectedArticle.FavoritesCount {
						t.Errorf(
							"Article[%d].FavoritesCount: got %d, want %d",
							i,
							article.FavoritesCount,
							expectedArticle.FavoritesCount,
						)
					}
					if article.Author.Username != expectedArticle.Author.Username {
						t.Errorf(
							"Article[%d].Author.Username: got %s, want %s",
							i,
							article.Author.Username,
							expectedArticle.Author.Username,
						)
					}
					if article.Author.Following != expectedArticle.Author.Following {
						t.Errorf(
							"Article[%d].Author.Following: got %t, want %t",
							i,
							article.Author.Following,
							expectedArticle.Author.Following,
						)
					}
				}
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				expected := tt.expectedResponse.(response.GenericErrorModel)
				if !reflect.DeepEqual(resp, expected) {
					t.Errorf("Response body: got %v, want %v", resp, expected)
				}
			}
		})
	}
}

// TestArticleHandler_GetArticlesFeed tests the GetArticlesFeed method of the ArticleHandler
func TestArticleHandler_GetArticlesFeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		queryParams      string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockArticleService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name:        "Successfully get articles feed",
			queryParams: "",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticlesFeedFunc: func(ctx context.Context, userID int64, limit, offset int) (*repository.ArticleListResult, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if limit != 20 {
							t.Errorf("Expected limit 20, got %d", limit)
						}
						if offset != 0 {
							t.Errorf("Expected offset 0, got %d", offset)
						}

						now := time.Now()
						articles := []*repository.Article{
							{
								ID:             1,
								Slug:           "feed-article-1",
								Title:          "Feed Article 1",
								Description:    "Feed Description 1",
								Body:           "Feed Body 1",
								TagList:        []string{"feed", "tag1"},
								CreatedAt:      now,
								UpdatedAt:      now,
								Favorited:      true,
								FavoritesCount: 3,
								Author: &repository.User{
									ID:        2,
									Username:  "followeduser",
									Bio:       "Followed User Bio",
									Image:     "https://example.com/followed.jpg",
									Following: true,
								},
							},
						}

						return &repository.ArticleListResult{
							Articles: articles,
							Count:    1,
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: MultipleArticlesResponse{
				Articles: []service.Article{
					{
						Slug:           "feed-article-1",
						Title:          "Feed Article 1",
						Description:    "Feed Description 1",
						Body:           "Feed Body 1",
						TagList:        []string{"feed", "tag1"},
						CreatedAt:      time.Time{},
						UpdatedAt:      time.Time{},
						Favorited:      true,
						FavoritesCount: 3,
						Author: service.Profile{
							Username:  "followeduser",
							Bio:       "Followed User Bio",
							Image:     "https://example.com/followed.jpg",
							Following: true,
						},
					},
				},
				ArticlesCount: 1,
			},
		},
		{
			name:        "Successfully get articles feed with pagination",
			queryParams: "?limit=5&offset=10",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticlesFeedFunc: func(ctx context.Context, userID int64, limit, offset int) (*repository.ArticleListResult, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if limit != 5 {
							t.Errorf("Expected limit 5, got %d", limit)
						}
						if offset != 10 {
							t.Errorf("Expected offset 10, got %d", offset)
						}

						return &repository.ArticleListResult{
							Articles: []*repository.Article{},
							Count:    0,
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: MultipleArticlesResponse{
				Articles:      []service.Article{},
				ArticlesCount: 0,
			},
		},
		{
			name:        "Unauthenticated request",
			queryParams: "",
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticlesFeedFunc: func(ctx context.Context, userID int64, limit, offset int) (*repository.ArticleListResult, error) {
						t.Errorf("GetArticlesFeed should not be called for unauthenticated request")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name:        "Internal server error",
			queryParams: "",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockArticleService {
				mockService := &MockArticleService{
					getArticlesFeedFunc: func(ctx context.Context, userID int64, limit, offset int) (*repository.ArticleListResult, error) {
						return nil, service.ErrInternalServer
					},
				}
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup Mock
			mockService := tt.setupMock()

			// Create Handler
			handler := NewArticleHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodGet,
				"/api/articles/feed"+tt.queryParams,
				nil,
			)

			// Setup Auth if provided
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.GetArticlesFeed()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			if tt.expectedStatus == http.StatusOK {
				var resp MultipleArticlesResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				for i := range resp.Articles {
					resp.Articles[i].CreatedAt = time.Time{}
					resp.Articles[i].UpdatedAt = time.Time{}
				}

				expected := tt.expectedResponse.(MultipleArticlesResponse)
				if resp.ArticlesCount != expected.ArticlesCount {
					t.Errorf(
						"ArticlesCount: got %d, want %d",
						resp.ArticlesCount,
						expected.ArticlesCount,
					)
				}
				if len(resp.Articles) != len(expected.Articles) {
					t.Errorf(
						"Articles length: got %d, want %d",
						len(resp.Articles),
						len(expected.Articles),
					)
				}
				for i, article := range resp.Articles {
					if i >= len(expected.Articles) {
						break
					}
					expectedArticle := expected.Articles[i]
					if article.Slug != expectedArticle.Slug {
						t.Errorf(
							"Article[%d].Slug: got %s, want %s",
							i,
							article.Slug,
							expectedArticle.Slug,
						)
					}
					if article.Title != expectedArticle.Title {
						t.Errorf(
							"Article[%d].Title: got %s, want %s",
							i,
							article.Title,
							expectedArticle.Title,
						)
					}
					if article.Favorited != expectedArticle.Favorited {
						t.Errorf(
							"Article[%d].Favorited: got %t, want %t",
							i,
							article.Favorited,
							expectedArticle.Favorited,
						)
					}
					if article.FavoritesCount != expectedArticle.FavoritesCount {
						t.Errorf(
							"Article[%d].FavoritesCount: got %d, want %d",
							i,
							article.FavoritesCount,
							expectedArticle.FavoritesCount,
						)
					}
					if article.Author.Username != expectedArticle.Author.Username {
						t.Errorf(
							"Article[%d].Author.Username: got %s, want %s",
							i,
							article.Author.Username,
							expectedArticle.Author.Username,
						)
					}
					if article.Author.Following != expectedArticle.Author.Following {
						t.Errorf(
							"Article[%d].Author.Following: got %t, want %t",
							i,
							article.Author.Following,
							expectedArticle.Author.Following,
						)
					}
				}
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				expected := tt.expectedResponse.(response.GenericErrorModel)
				if !reflect.DeepEqual(resp, expected) {
					t.Errorf("Response body: got %v, want %v", resp, expected)
				}
			}
		})
	}
}
