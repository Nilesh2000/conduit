package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"conduit/internal/middleware"
	"conduit/internal/response"
	"conduit/internal/service"
)

// MockProfileService is a mock implementation of the ProfileService interface
type MockProfileService struct {
	getProfileFunc   func(ctx context.Context, username string, currentUserID int64) (*service.Profile, error)
	followUserFunc   func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error)
	unfollowUserFunc func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error)
}

// GetProfile returns a mock profile
func (m *MockProfileService) GetProfile(
	ctx context.Context,
	username string,
	currentUserID int64,
) (*service.Profile, error) {
	return m.getProfileFunc(ctx, username, currentUserID)
}

// FollowUser returns a mock profile
func (m *MockProfileService) FollowUser(
	ctx context.Context,
	followerID int64,
	followingName string,
) (*service.Profile, error) {
	return m.followUserFunc(ctx, followerID, followingName)
}

// UnfollowUser returns a mock profile
func (m *MockProfileService) UnfollowUser(
	ctx context.Context,
	followerID int64,
	followingName string,
) (*service.Profile, error) {
	return m.unfollowUserFunc(ctx, followerID, followingName)
}

// Test_profileHandler_GetProfile tests the GetProfile method of the profileHandler
func Test_profileHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		setupAuth        func(r *http.Request)
		setupMock        func() *MockProfileService
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name:     "Profile found (authenticated)",
			username: "testuser",
			setupAuth: func(r *http.Request) {
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				*r = *r.WithContext(ctx)
				r.Header.Set("Authorization", "Token testtoken")
			},
			setupMock: func() *MockProfileService {
				mockService := &MockProfileService{
					getProfileFunc: func(ctx context.Context, username string, currentUserID int64) (*service.Profile, error) {
						if username != "testuser" {
							t.Errorf("Expected username 'testuser', got %q", username)
						}
						if currentUserID != 1 {
							t.Errorf("Expected currentUserID 1, got %d", currentUserID)
						}

						return &service.Profile{
							Username:  "testuser",
							Bio:       "Test bio",
							Image:     "https://example.com/image.jpg",
							Following: false,
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ProfileResponse{
				Profile: service.Profile{
					Username:  "testuser",
					Bio:       "Test bio",
					Image:     "https://example.com/image.jpg",
					Following: false,
				},
			},
		},
		{
			name:     "Profile found (unauthenticated)",
			username: "testuser",
			setupAuth: func(r *http.Request) {
				// Do not add user ID to context
			},
			setupMock: func() *MockProfileService {
				mockService := &MockProfileService{
					getProfileFunc: func(ctx context.Context, username string, currentUserID int64) (*service.Profile, error) {
						if username != "testuser" {
							t.Errorf("Expected username 'testuser', got %q", username)
						}
						if currentUserID != 0 {
							t.Errorf("Expected currentUserID 0, got %d", currentUserID)
						}

						return &service.Profile{
							Username:  "testuser",
							Bio:       "Test bio",
							Image:     "https://example.com/image.jpg",
							Following: false,
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ProfileResponse{
				Profile: service.Profile{
					Username:  "testuser",
					Bio:       "Test bio",
					Image:     "https://example.com/image.jpg",
					Following: false,
				},
			},
		},
		{
			name:     "Profile not found",
			username: "nonexistentuser",
			setupAuth: func(r *http.Request) {
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				*r = *r.WithContext(ctx)
				r.Header.Set("Authorization", "Token testtoken")
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					getProfileFunc: func(ctx context.Context, username string, currentUserID int64) (*service.Profile, error) {
						return nil, service.ErrUserNotFound
					},
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"User not found"}},
			},
		},
		{
			name:     "Internal server error",
			username: "testuser",
			setupAuth: func(r *http.Request) {
				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				*r = *r.WithContext(ctx)
				r.Header.Set("Authorization", "Token testtoken")
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					getProfileFunc: func(ctx context.Context, username string, currentUserID int64) (*service.Profile, error) {
						return nil, errors.New("internal server error")
					},
				}
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

			// Setup mock service
			mockService := tt.setupMock()

			// Create handler
			profileHandler := NewProfileHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/profiles/"+tt.username, nil)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				tt.setupAuth(req)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := profileHandler.GetProfile()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got interface{}
			if tt.expectedStatus == http.StatusOK {
				var resp ProfileResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			} else {
				var resp response.GenericErrorModel
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			}

			if !reflect.DeepEqual(got, tt.expectedResponse) {
				t.Errorf("Response body: got %v, want %v", got, tt.expectedResponse)
			}
		})
	}
}
