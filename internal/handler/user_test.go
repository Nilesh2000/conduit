package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
)

// MockUserService is a mock implementation of the UserService interface
type MockUserService struct {
	registerFunc       func(ctx context.Context, username, email, password string) (*service.User, error)
	loginFunc          func(ctx context.Context, email, password string) (*service.User, error)
	getCurrentUserFunc func(ctx context.Context, userID int64) (*service.User, error)
	updateUserFunc     func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error)
}

// Ensure MockUserService implements the UserService interface
var _ UserService = (*MockUserService)(nil)

// Register creates a new user in the mock service
func (m *MockUserService) Register(
	ctx context.Context,
	username, email, password string,
) (*service.User, error) {
	return m.registerFunc(ctx, username, email, password)
}

// Login logs in a user in the mock service
func (m *MockUserService) Login(
	ctx context.Context,
	email, password string,
) (*service.User, error) {
	return m.loginFunc(ctx, email, password)
}

// GetCurrentUser gets the current user in the mock service
func (m *MockUserService) GetCurrentUser(ctx context.Context, userID int64) (*service.User, error) {
	return m.getCurrentUserFunc(ctx, userID)
}

// UpdateUser updates a user in the mock service
func (m *MockUserService) UpdateUser(
	ctx context.Context,
	userID int64,
	username, email, password, bio, image *string,
) (*service.User, error) {
	return m.updateUserFunc(ctx, userID, username, email, password, bio, image)
}

// TestUserHandler_Register tests the Register method of the UserHandler
func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestBody      string
		setupMock        func() *MockUserService
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "Valid registration",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						if username != "testuser" || email != "test@example.com" ||
							password != "password123" {
							t.Errorf(
								"Expected Register(%q, %q, %q), got Register(%q, %q, %q)",
								"testuser",
								"test@example.com",
								"password123",
								username,
								email,
								password,
							)
						}
						return &service.User{
							Email:    email,
							Token:    "jwt.token.here",
							Username: username,
							Bio:      "",
							Image:    "",
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: UserResponse{
				User: service.User{
					Email:    "test@example.com",
					Token:    "jwt.token.here",
					Username: "testuser",
					Bio:      "",
					Image:    "",
				},
			},
		},
		{
			name: "Invalid JSON",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "password123",
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						t.Errorf("Register should not be called for invalid JSON")
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
			requestBody: `{
				"user": {
					"email": "test@example.com"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						t.Errorf("Register should not be called for missing required fields")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Username is required", "Password is required"}},
			},
		},
		{
			name: "Invalid email",
			requestBody: `{
					"user": {
						"username": "testuser",
						"email": "invalid-email",
						"password": "password123"
					}
				}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						t.Errorf("Register should not be called for invalid email")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"invalid-email is not a valid email"}},
			},
		},
		{
			name: "Password too short",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "short"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						t.Errorf("Register should not be called for short password")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Password must be at least 8 characters long"}},
			},
		},
		{
			name: "Username already taken",
			requestBody: `{
				"user": {
					"username": "existinguser",
					"email": "test@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						return nil, service.ErrUsernameTaken
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Username already taken"}},
			},
		},
		{
			name: "Email already registered",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "existing@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
						return nil, service.ErrEmailTaken
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Email already registered"}},
			},
		},
		{
			name: "Internal server error",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					registerFunc: func(ctx context.Context, username, email, password string) (*service.User, error) {
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

			// Setup mock service
			mockUserService := tt.setupMock()

			// Create handler
			userHandler := NewUserHandler(mockUserService)

			// Create request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/users",
				strings.NewReader(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := userHandler.Register()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got interface{}
			if tt.expectedStatus == http.StatusCreated {
				var resp UserResponse
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

// TestUserHandler_Login tests the Login method of the UserHandler
func TestUserHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestBody      string
		setupMock        func() *MockUserService
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "Valid login",
			requestBody: `{
				"user": {
					"email": "test@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					loginFunc: func(ctx context.Context, email, password string) (*service.User, error) {
						if email != "test@example.com" || password != "password123" {
							t.Errorf(
								"Expected Login(%q, %q), got Login(%q, %q)",
								"test@example.com",
								"password123",
								email,
								password,
							)
						}
						return &service.User{
							Email:    email,
							Token:    "jwt.token.here",
							Username: "testuser",
							Bio:      "I'm a test user",
							Image:    "https://example.com/image.jpg",
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: UserResponse{
				User: service.User{
					Email:    "test@example.com",
					Token:    "jwt.token.here",
					Username: "testuser",
					Bio:      "I'm a test user",
					Image:    "https://example.com/image.jpg",
				},
			},
		},
		{
			name: "Invalid JSON",
			requestBody: `{
				"user": {
					"email": "test@example.com",
					"password": "password123",
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					loginFunc: func(ctx context.Context, email, password string) (*service.User, error) {
						t.Errorf("Login should not be called for invalid JSON")
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
			requestBody: `{
				"user": {
					"email": "test@example.com"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					loginFunc: func(ctx context.Context, email, password string) (*service.User, error) {
						t.Errorf("Login should not be called for missing required fields")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Password is required"}},
			},
		},
		{
			name: "Invalid credentials",
			requestBody: `{
				"user": {
					"email": "test@example.com",
					"password": "wrongpassword"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					loginFunc: func(ctx context.Context, email, password string) (*service.User, error) {
						return nil, service.ErrInvalidCredentials
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Invalid credentials"}},
			},
		},
		{
			name: "User not found",
			requestBody: `{
				"user": {
					"email": "nonexistent@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					loginFunc: func(ctx context.Context, email, password string) (*service.User, error) {
						return nil, service.ErrUserNotFound
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Invalid credentials"}},
			},
		},
		{
			name: "Internal server error",
			requestBody: `{
				"user": {
					"email": "test@example.com",
					"password": "password123"
				}
			}`,
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					loginFunc: func(ctx context.Context, email, password string) (*service.User, error) {
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

			// Setup mock service
			mockUserService := tt.setupMock()

			// Create handler
			userHandler := NewUserHandler(mockUserService)

			// Create request
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/users/login",
				strings.NewReader(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := userHandler.Login()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got interface{}
			if tt.expectedStatus == http.StatusOK {
				var resp UserResponse
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

// TestUserHandler_GetCurrentUser tests the GetCurrentUser method of the UserHandler
func TestUserHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockUserService
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "Valid current user",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					getCurrentUserFunc: func(ctx context.Context, userID int64) (*service.User, error) {
						if userID != 1 {
							t.Errorf("Expected service called with userID 1, got %d", userID)
						}
						return &service.User{
							Email:    "test@example.com",
							Username: "testuser",
							Bio:      "I'm a test user",
							Image:    "https://example.com/image.jpg",
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: UserResponse{
				User: service.User{
					Email:    "test@example.com",
					Token:    "jwt.token.here",
					Username: "testuser",
					Bio:      "I'm a test user",
					Image:    "https://example.com/image.jpg",
				},
			},
		},
		{
			name: "Unauthenticated request",
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					getCurrentUserFunc: func(ctx context.Context, userID int64) (*service.User, error) {
						t.Errorf("Service should not be called for unauthenticated request")
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
			name: "User not found",
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					getCurrentUserFunc: func(ctx context.Context, userID int64) (*service.User, error) {
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
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					getCurrentUserFunc: func(ctx context.Context, userID int64) (*service.User, error) {
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

			// Setup mock service
			mockUserService := tt.setupMock()

			// Create handler
			userHandler := NewUserHandler(mockUserService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/user", nil)

			// Add authorization and setup context
			req = tt.setupAuth(req)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := userHandler.GetCurrentUser()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got interface{}
			if tt.expectedStatus == http.StatusOK {
				var resp UserResponse
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

func TestUserHandler_UpdateCurrentUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		requestBody      string
		setupAuth        func(r *http.Request) *http.Request
		setupMock        func() *MockUserService
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "Valid update",
			requestBody: `{
				"user": {
					"username": "updateduser",
					"email": "updated@example.com",
					"password": "newpassword123",
					"bio": "Updated bio",
					"image": "https://example.com/updated.jpg"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						if userID != 1 {
							t.Errorf("Expected service called with userID 1, got %d", userID)
						}
						if *username != "updateduser" {
							t.Errorf("Expected username 'updateduser', got %q", *username)
						}
						if *email != "updated@example.com" {
							t.Errorf("Expected email 'updated@example.com', got %q", *email)
						}
						if *password != "newpassword123" {
							t.Errorf("Expected password 'newpassword123', got %q", *password)
						}
						if *bio != "Updated bio" {
							t.Errorf("Expected bio 'Updated bio', got %q", *bio)
						}
						if *image != "https://example.com/updated.jpg" {
							t.Errorf(
								"Expected image 'https://example.com/updated.jpg', got %q",
								*image,
							)
						}
						return &service.User{
							Email:    *email,
							Username: *username,
							Bio:      *bio,
							Image:    *image,
							// Token is not set in the service response
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: UserResponse{
				User: service.User{
					Email:    "updated@example.com",
					Token:    "jwt.token.here",
					Username: "updateduser",
					Bio:      "Updated bio",
					Image:    "https://example.com/updated.jpg",
				},
			},
		},
		{
			name: "Partial update - only email",
			requestBody: `{
				"user": {
					"email": "newmail@example.com"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						if *email != "newmail@example.com" {
							t.Errorf("Expected email 'newmail@example.com', got %q", *email)
						}

						// Other fields should be unchanged
						if username != nil {
							t.Errorf("Expected username to be unchanged, got %q", *username)
						}

						return &service.User{
							Email:    *email,
							Username: "existinguser", // Assume service returns existing username
							Bio:      "Existing bio",
							Image:    "https://example.com/existing.jpg",
							// Token is not set in the service response
						}, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusOK,
			expectedResponse: UserResponse{
				User: service.User{
					Email:    "newmail@example.com",
					Token:    "jwt.token.here",
					Username: "existinguser",
					Bio:      "Existing bio",
					Image:    "https://example.com/existing.jpg",
				},
			},
		},
		{
			name: "Unauthenticated request",
			requestBody: `{
				"user": {
					"email": "newmail@example.com"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				// Don't add user ID to context to simulate unauthenticated request
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						t.Errorf("Service should not be called for unauthenticated request")
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
				"user": {
					"email": "newmail@example.com",
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						t.Errorf("Service should not be called for invalid request JSON")
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
			name: "Invalid email",
			requestBody: `{
				"user": {
					"email": "invalid-email"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						t.Errorf("Service should not be called for invalid email")
						return nil, nil
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"invalid-email is not a valid email"}},
			},
		},
		{
			name: "User not found",
			requestBody: `{
				"user": {
					"username": "updateduser"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
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
			name: "Username already taken",
			requestBody: `{
				"user": {
					"username": "existinguser"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						return nil, service.ErrUsernameTaken
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Username already taken"}},
			},
		},
		{
			name: "Email already registered",
			requestBody: `{
				"user": {
					"email": "existing@example.com"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
						return nil, service.ErrEmailTaken
					},
				}
				return mockService
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: response.GenericErrorModel{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"Email already registered"}},
			},
		},
		{
			name: "Internal server error",
			requestBody: `{
				"user": {
					"username": "updateduser",
					"email": "updated@example.com",
					"password": "newpassword123",
					"bio": "Updated bio",
					"image": "https://example.com/updated.jpg"
				}
			}`,
			setupAuth: func(r *http.Request) *http.Request {
				r.Header.Set("Authorization", "Token jwt.token.here")

				ctx := r.Context()
				ctx = context.WithValue(ctx, middleware.UserIDContextKey, int64(1))
				r = r.WithContext(ctx)
				return r
			},
			setupMock: func() *MockUserService {
				mockService := &MockUserService{
					updateUserFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*service.User, error) {
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

			// Setup mock service
			mockUserService := tt.setupMock()

			// Create handler
			userHandler := NewUserHandler(mockUserService)

			// Create request
			req := httptest.NewRequest(
				http.MethodPut,
				"/api/user",
				strings.NewReader(tt.requestBody),
			)
			req.Header.Set("Content-Type", "application/json")

			// Add authorization token and setup context
			req = tt.setupAuth(req)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve request
			handler := userHandler.UpdateCurrentUser()
			handler.ServeHTTP(rr, req)

			// Check status code
			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			// Check response body
			var got interface{}
			if tt.expectedStatus == http.StatusOK {
				var resp UserResponse
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
