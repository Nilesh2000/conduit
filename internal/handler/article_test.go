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
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
)

// MockArticleService is a mock implementation of the ArticleService interface
type MockArticleService struct {
	createArticleFunc     func(ctx context.Context, userID int64, title, description, body string, tagList []string) (*service.Article, error)
	getArticleFunc        func(ctx context.Context, slug string) (*service.Article, error)
	favoriteArticleFunc   func(ctx context.Context, userID int64, slug string) (*service.Article, error)
	unfavoriteArticleFunc func(ctx context.Context, userID int64, slug string) (*service.Article, error)
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
) (*service.Article, error) {
	return m.getArticleFunc(ctx, slug)
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
					getArticleFunc: func(ctx context.Context, slug string) (*service.Article, error) {
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
					getArticleFunc: func(ctx context.Context, slug string) (*service.Article, error) {
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
					getArticleFunc: func(ctx context.Context, slug string) (*service.Article, error) {
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
