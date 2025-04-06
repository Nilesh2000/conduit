package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

type MockUserService struct {
	RegisterFunc func(username, email, password string) (*User, error)
}

var _ UserService = (*MockUserService)(nil)

func (m *MockUserService) Register(username, email, password string) (*User, error) {
	return m.RegisterFunc(username, email, password)
}

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		mockRegister     func(username, email, password string) (*User, error)
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
			mockRegister: func(username, email, password string) (*User, error) {
				if username != "testuser" || email != "test@example.com" || password != "password" {
					t.Errorf("Expected Register(%q, %q, %q), got Register(%q, %q, %q)", "testuser", "test@example.com", "password", username, email, password)
				}
				return &User{
					Email:    email,
					Token:    "jwt.token.here",
					Username: username,
					Bio:      "",
					Image:    "",
				}, nil
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: UserResponse{
				User: User{
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
			mockRegister: func(username, email, password string) (*User, error) {
				t.Errorf("Register should not be called for invalid JSON")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
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
			mockRegister: func(username, email, password string) (*User, error) {
				t.Errorf("Register should not be called for missing required fields")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
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
			mockRegister: func(username, email, password string) (*User, error) {
				t.Errorf("Register should not be called for invalid email")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
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
			mockRegister: func(username, email, password string) (*User, error) {
				t.Errorf("Register should not be called for short password")
				return nil, nil
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
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
			mockRegister: func(username, email, password string) (*User, error) {
				return nil, ErrUsernameTaken
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
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
			mockRegister: func(username, email, password string) (*User, error) {
				return nil, ErrEmailTaken
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
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
			mockRegister: func(username, email, password string) (*User, error) {
				return nil, ErrInternalServer
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: ErrorResponse{
				Errors: struct {
					Body []string
				}{Body: []string{"internal server error"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := &MockUserService{
				RegisterFunc: tt.mockRegister,
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

type MockUserRepository struct {
	CreateUserFunc        func(username, email, password string) (*UserRepo, error)
	GetUserByUsernameFunc func(username string) (*UserRepo, error)
	GetUserByEmailFunc    func(email string) (*UserRepo, error)
}

var _ UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) CreateUser(username, email, password string) (*UserRepo, error) {
	return m.CreateUserFunc(username, email, password)
}

func (m *MockUserRepository) GetUserByUsername(username string) (*UserRepo, error) {
	return m.GetUserByUsernameFunc(username)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*UserRepo, error) {
	return m.GetUserByEmailFunc(email)
}

func Test_userService_Register(t *testing.T) {
	jwtSecret := "test-secret"
	jwtExpiration := time.Hour * 24

	tests := []struct {
		name           string
		username       string
		email          string
		password       string
		mockCreateUser func(username, email, password string) (*UserRepo, error)
		expectedError  error
		validateFunc   func(*User)
	}{
		{
			name:     "Valid registration",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			mockCreateUser: func(username, email, password string) (*UserRepo, error) {
				if username != "testuser" || email != "test@example.com" {
					t.Errorf("Expected CreateUser(%q, %q, _), got CreateUser(%q, %q, _)", "testuser", "test@example.com", username, email)
				}

				// Verify password is hashed
				if password == "password" {
					t.Errorf("Password should be hashed, got plain text")
				}

				return &UserRepo{
					ID:       1,
					Username: username,
					Email:    email,
					Password: password,
					Bio:      "",
					Image:    "",
				}, nil
			},
			expectedError: nil,
			validateFunc: func(u *User) {
				if u.Username != "testuser" {
					t.Errorf("Expected Username 'testuser', got %q", u.Username)
				}
				if u.Email != "test@example.com" {
					t.Errorf("Expected Email 'test@example.com', got %q", u.Email)
				}
				if u.Token == "" {
					t.Errorf("Expected token to be non-empty")
				}

				// Verify token is a valid JWT
				token, err := jwt.Parse(u.Token, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				if !token.Valid || err != nil {
					t.Errorf("Invalid token: %v", err)
				}

				// Verify claims
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if claims["id"] != float64(1) {
						t.Errorf("Expected token claim ID to be 1, got %q", claims["id"])
					}
				} else {
					t.Errorf("Failed to parse token claims")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepository := &MockUserRepository{
				CreateUserFunc: tt.mockCreateUser,
			}

			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			user, err := userService.Register(tt.username, tt.email, tt.password)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if err == nil && tt.validateFunc != nil {
				tt.validateFunc(user)
			}
		})
	}
}
