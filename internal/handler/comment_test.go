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

// MockCommentService is a mock implementation of the CommentService interface
type MockCommentService struct {
	getCommentsFunc   func(ctx context.Context, slug string, userID *int64) ([]service.Comment, error)
	createCommentFunc func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error)
	deleteCommentFunc func(ctx context.Context, userID int64, slug string, commentID int64) error
}

// GetComments gets comments for an article in the mock service
func (m *MockCommentService) GetComments(
	ctx context.Context,
	slug string,
	userID *int64,
) ([]service.Comment, error) {
	return m.getCommentsFunc(ctx, slug, userID)
}

// CreateComment creates a comment in the mock service
func (m *MockCommentService) CreateComment(
	ctx context.Context,
	userID int64,
	slug string,
	body string,
) (*service.Comment, error) {
	return m.createCommentFunc(ctx, userID, slug, body)
}

// DeleteComment deletes a comment in the mock service
func (m *MockCommentService) DeleteComment(
	ctx context.Context,
	userID int64,
	slug string,
	commentID int64,
) error {
	return m.deleteCommentFunc(ctx, userID, slug, commentID)
}

// TestCommentHandler_GetComments tests the GetComments method of the CommentHandler
func TestCommentHandler_GetComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockCommentService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Successfully get comments",
			slug: "test-slug",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					getCommentsFunc: func(ctx context.Context, slug string, userID *int64) ([]service.Comment, error) {
						if slug != "test-slug" {
							t.Errorf("Expected slug 'test-slug', got %q", slug)
						}

						now := time.Now()
						return []service.Comment{
							{
								ID:        1,
								Body:      "This is a test comment",
								CreatedAt: now,
								UpdatedAt: now,
								Author: service.Profile{
									Username:  "testuser",
									Bio:       "test bio",
									Image:     "test image",
									Following: false,
								},
							},
							{
								ID:        2,
								Body:      "This is a second test comment",
								CreatedAt: now,
								UpdatedAt: now,
								Author: service.Profile{
									Username:  "testuser",
									Bio:       "test bio",
									Image:     "test image",
									Following: false,
								},
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: CommentsResponse{
				Comments: []service.Comment{
					{
						ID:        1,
						Body:      "This is a test comment",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
						Author: service.Profile{
							Username:  "testuser",
							Bio:       "test bio",
							Image:     "test image",
							Following: false,
						},
					},
					{
						ID:        2,
						Body:      "This is a second test comment",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
						Author: service.Profile{
							Username:  "testuser",
							Bio:       "test bio",
							Image:     "test image",
							Following: false,
						},
					},
				},
			},
		},
		{
			name: "Unauthenticated request",
			slug: "test-slug",
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					getCommentsFunc: func(ctx context.Context, slug string, userID *int64) ([]service.Comment, error) {
						if slug != "test-slug" {
							t.Errorf("Expected slug 'test-slug', got %q", slug)
						}

						if userID != nil {
							t.Errorf("Expected userID nil, got %d", *userID)
						}

						now := time.Now()
						return []service.Comment{
							{
								ID:        1,
								Body:      "This is a test comment",
								CreatedAt: now,
								UpdatedAt: now,
								Author: service.Profile{
									Username:  "testuser",
									Bio:       "test bio",
									Image:     "test image",
									Following: false,
								},
							},
							{
								ID:        2,
								Body:      "This is a second test comment",
								CreatedAt: now,
								UpdatedAt: now,
								Author: service.Profile{
									Username:  "testuser",
									Bio:       "test bio",
									Image:     "test image",
									Following: false,
								},
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: CommentsResponse{
				Comments: []service.Comment{
					{
						ID:        1,
						Body:      "This is a test comment",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
						Author: service.Profile{
							Username:  "testuser",
							Bio:       "test bio",
							Image:     "test image",
							Following: false,
						},
					},
					{
						ID:        2,
						Body:      "This is a second test comment",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
						Author: service.Profile{
							Username:  "testuser",
							Bio:       "test bio",
							Image:     "test image",
							Following: false,
						},
					},
				},
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
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					getCommentsFunc: func(ctx context.Context, slug string, userID *int64) ([]service.Comment, error) {
						if slug != "non-existent-article" {
							t.Errorf("Expected slug 'non-existent-article', got %q", slug)
						}

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
			slug: "test-slug",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					getCommentsFunc: func(ctx context.Context, slug string, userID *int64) ([]service.Comment, error) {
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
			handler := NewCommentHandler(mockService)

			// Create Request
			req := httptest.NewRequest(http.MethodGet, "/api/articles/"+tt.slug+"/comments", nil)
			req.SetPathValue("slug", tt.slug)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.GetComments()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp CommentsResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				for i := range resp.Comments {
					resp.Comments[i].CreatedAt = time.Time{}
					resp.Comments[i].UpdatedAt = time.Time{}
				}
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

// TestCommentHandler_CreateComment tests the CreateComment method of the CommentHandler
func TestCommentHandler_CreateComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		requestBody      any
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockCommentService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "Valid comment creation",
			slug: "test-slug",
			requestBody: NewComment{
				Comment: struct {
					Body string `json:"body" validate:"required"`
				}{
					Body: "This is a test comment",
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					createCommentFunc: func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error) {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if slug != "test-slug" {
							t.Errorf("Expected slug 'test-slug', got %q", slug)
						}
						if body != "This is a test comment" {
							t.Errorf("Expected body 'This is a test comment', got %q", body)
						}

						now := time.Now()
						return &service.Comment{
							ID:        1,
							Body:      body,
							CreatedAt: now,
							UpdatedAt: now,
							Author: service.Profile{
								Username:  "testuser",
								Bio:       "test bio",
								Image:     "test image",
								Following: false,
							},
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: CommentResponse{
				Comment: service.Comment{
					ID:        1,
					Body:      "This is a test comment",
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
					Author: service.Profile{
						Username:  "testuser",
						Bio:       "test bio",
						Image:     "test image",
						Following: false,
					},
				},
			},
		},
		{
			name: "Unauthenticated request",
			slug: "test-slug",
			requestBody: NewComment{
				Comment: struct {
					Body string `json:"body" validate:"required"`
				}{
					Body: "This is a test comment",
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					createCommentFunc: func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error) {
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
			name:        "Invalid request body",
			slug:        "test-slug",
			requestBody: "invalid-request-body",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					createCommentFunc: func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error) {
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Invalid request body"}},
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			requestBody: NewComment{
				Comment: struct {
					Body string `json:"body" validate:"required"`
				}{
					Body: "This is a test comment",
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					createCommentFunc: func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error) {
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
			slug: "test-slug",
			requestBody: NewComment{
				Comment: struct {
					Body string `json:"body" validate:"required"`
				}{
					Body: "This is a test comment",
				},
			},
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					createCommentFunc: func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error) {
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
			handler := NewCommentHandler(mockService)

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
				"/api/articles/"+tt.slug+"/comments",
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
			handler.CreateComment()(rr, req)

			// Check Status Code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Status code: got %v, want %v", status, tt.expectedStatus)
			}

			// Check Response Body
			var got any
			if tt.expectedStatus == http.StatusOK {
				var resp CommentResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				// Reset timestamps for comparison
				resp.Comment.CreatedAt = time.Time{}
				resp.Comment.UpdatedAt = time.Time{}
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

// TestCommentHandler_DeleteComment tests the DeleteComment method of the CommentHandler
func TestCommentHandler_DeleteComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		slug             string
		commentID        string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockCommentService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name:      "Successfully delete comment",
			slug:      "test-slug",
			commentID: "1",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					deleteCommentFunc: func(ctx context.Context, userID int64, slug string, commentID int64) error {
						if userID != 1 {
							t.Errorf("Expected userID 1, got %d", userID)
						}
						if slug != "test-slug" {
							t.Errorf("Expected slug 'test-slug', got %q", slug)
						}
						if commentID != 1 {
							t.Errorf("Expected commentID 1, got %d", commentID)
						}

						return nil
					},
				}
				return mockService
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: nil,
		},
		{
			name:      "Unauthenticated request",
			slug:      "test-slug",
			commentID: "1",
			setupAuth: func(r *http.Request) *http.Request {
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					deleteCommentFunc: func(ctx context.Context, userID int64, slug string, commentID int64) error {
						t.Errorf("DeleteComment should not be called for unauthenticated request")
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
			name:      "Invalid comment ID",
			slug:      "test-slug",
			commentID: "invalid-comment-id",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					deleteCommentFunc: func(ctx context.Context, userID int64, slug string, commentID int64) error {
						t.Errorf("DeleteComment should not be called for invalid comment ID")
						return nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Invalid comment ID"}},
			},
		},
		{
			name:      "Not the author of the comment",
			slug:      "test-slug",
			commentID: "1",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(2))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					deleteCommentFunc: func(ctx context.Context, userID int64, slug string, commentID int64) error {
						return service.ErrCommentNotAuthorized
					},
				}
				return mockService
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"You are not the author of this comment"}},
			},
		},
		{
			name:      "Comment not found",
			slug:      "test-slug",
			commentID: "1",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					deleteCommentFunc: func(ctx context.Context, userID int64, slug string, commentID int64) error {
						return service.ErrCommentNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Comment not found"}},
			},
		},
		{
			name:      "Internal server error",
			slug:      "test-slug",
			commentID: "1",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockCommentService {
				mockService := &MockCommentService{
					deleteCommentFunc: func(ctx context.Context, userID int64, slug string, commentID int64) error {
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
			handler := NewCommentHandler(mockService)

			// Create Request
			req := httptest.NewRequest(
				http.MethodDelete,
				"/api/articles/"+tt.slug+"/comments/"+tt.commentID,
				nil,
			)
			req.SetPathValue("slug", tt.slug)
			req.SetPathValue("id", tt.commentID)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create Response Recorder
			rr := httptest.NewRecorder()

			// Serve Request
			handler.DeleteComment()(rr, req)

			// Check Status Code
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
