package handler

import (
	"conduit/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type MockUserService struct {
	registerFunc func(username, email, password string) (*service.User, error)
}

var _ UserService = (*MockUserService)(nil)

func (m *MockUserService) Register(username, email, password string) (*service.User, error) {
	return m.registerFunc(username, email, password)
}

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		mockRegister     func(username, email, password string) (*service.User, error)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "Valid registration",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "password"
				}
			}`,
			mockRegister: func(username, email, password string) (*service.User, error) {
				if username != "testuser" || email != "test@example.com" || password != "password" {
					t.Errorf("Expected Register(%q, %q, %q), got Register(%q, %q, %q)", "testuser", "test@example.com", "password", username, email, password)
				}
				return &service.User{
					Email:    email,
					Token:    "jwt.token.here",
					Username: username,
					Bio:      "",
					Image:    "",
				}, nil
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
			}`,
			mockRegister: func(username, email, password string) (*service.User, error) {
				t.Errorf("Register should not be called for invalid JSON")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
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
			mockRegister: func(username, email, password string) (*service.User, error) {
				t.Errorf("Register should not be called for missing required fields")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
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
						"password": "password"
					}
				}`,
			mockRegister: func(username, email, password string) (*service.User, error) {
				t.Errorf("Register should not be called for invalid email")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
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
			mockRegister: func(username, email, password string) (*service.User, error) {
				t.Errorf("Register should not be called for short password")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
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
					"password": "password"
				}
			}`,
			mockRegister: func(username, email, password string) (*service.User, error) {
				return nil, service.ErrUsernameTaken
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"username already taken"}},
			},
		},
		{
			name: "Email already registered",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "existing@example.com",
					"password": "password"
				}
			}`,
			mockRegister: func(username, email, password string) (*service.User, error) {
				return nil, service.ErrEmailTaken
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"email already registered"}},
			},
		},
		{
			name: "Internal server error",
			requestBody: `{
				"user": {
					"username": "testuser",
					"email": "test@example.com",
					"password": "password"
				}
			}`,
			mockRegister: func(username, email, password string) (*service.User, error) {
				return nil, service.ErrInternalServer
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string `json:"body"`
				}{Body: []string{"internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := &MockUserService{
				registerFunc: tt.mockRegister,
			}

			userHandler := NewUserHandler(mockUserService)

			req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler := userHandler.Register()
			handler.ServeHTTP(rr, req)

			if got, want := rr.Code, tt.expectedStatus; got != want {
				t.Errorf("Status code: got %v, want %v", got, want)
			}

			var got interface{}
			if tt.expectedStatus == http.StatusCreated {
				var resp UserResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				got = resp
			} else {
				var resp ErrorResponse
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
