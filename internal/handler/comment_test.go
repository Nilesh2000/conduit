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
	createCommentFunc func(ctx context.Context, userID int64, slug string, body string) (*service.Comment, error)
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
