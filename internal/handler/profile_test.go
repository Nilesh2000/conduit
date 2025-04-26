package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
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
	t.Parallel()

	tests := []struct {
		name             string
		username         string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockProfileService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name:     "Profile found (authenticated)",
			username: "testuser",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
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
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
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
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
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
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
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
				req = tt.setupAuth(req)
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
			var got any
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

// Test_profileHandler_Follow tests the Follow method of the profileHandler
func Test_profileHandler_Follow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		username         string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockProfileService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name:     "Successfully follow user",
			username: "usertofollow",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						if followerID != 1 {
							t.Errorf("Expected followerID 1, got %d", followerID)
						}
						if followingName != "usertofollow" {
							t.Errorf("Expected followingName 'usertofollow', got %q", followingName)
						}

						return &service.Profile{
							Username:  "usertofollow",
							Bio:       "Their bio",
							Image:     "https://example.com/their-image.jpg",
							Following: true,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ProfileResponse{
				Profile: service.Profile{
					Username:  "usertofollow",
					Bio:       "Their bio",
					Image:     "https://example.com/their-image.jpg",
					Following: true,
				},
			},
		},
		{
			name:     "Unauthenticated User",
			username: "usertofollow",
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						t.Error("Service should not be called for unauthenticated request")
						return nil, nil
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name:     "User not found",
			username: "nonexistentuser",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
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
			name:     "Cannot follow self",
			username: "currentuser",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						return nil, service.ErrCannotFollowSelf
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Cannot follow yourself"}},
			},
		},
		{
			name:     "Internal server error",
			username: "usertofollow",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					followUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						return nil, service.ErrInternalServer
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
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+tt.username+"/follow", nil)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := profileHandler.Follow()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got any
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

// Test_profileHandler_Unfollow tests the Unfollow method of the profileHandler
func Test_profileHandler_Unfollow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		username         string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockProfileService
		expectedStatus   int
		expectedResponse any
	}{
		{
			name:     "Successfully unfollow user",
			username: "usertounfollow",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						if followerID != 1 {
							t.Errorf("Expected followerID 1, got %d", followerID)
						}
						if followingName != "usertounfollow" {
							t.Errorf(
								"Expected followingName 'usertounfollow', got %q",
								followingName,
							)
						}

						return &service.Profile{
							Username:  "usertounfollow",
							Bio:       "Their bio",
							Image:     "https://example.com/their-image.jpg",
							Following: false,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedResponse: ProfileResponse{
				Profile: service.Profile{
					Username:  "usertounfollow",
					Bio:       "Their bio",
					Image:     "https://example.com/their-image.jpg",
					Following: false,
				},
			},
		},
		{
			name:     "Unauthenticated User",
			username: "usertounfollow",
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						t.Error("Service should not be called for unauthenticated request")
						return nil, nil
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Unauthorized"}},
			},
		},
		{
			name:     "User not found",
			username: "nonexistentuser",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
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
			name:     "Cannot unfollow self",
			username: "currentuser",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						return nil, service.ErrCannotFollowSelf
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Cannot unfollow yourself"}},
			},
		},
		{
			name:     "Internal server error",
			username: "usertounfollow",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token testtoken")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockProfileService {
				return &MockProfileService{
					unfollowUserFunc: func(ctx context.Context, followerID int64, followingName string) (*service.Profile, error) {
						return nil, service.ErrInternalServer
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
			req := httptest.NewRequest(http.MethodPost, "/api/profiles/"+tt.username+"/follow", nil)

			// Add authorization token and setup context
			if tt.setupAuth != nil {
				req = tt.setupAuth(req)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := profileHandler.Unfollow()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got any
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
