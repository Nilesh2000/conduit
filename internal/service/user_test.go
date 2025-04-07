package service

import (
	"conduit/internal/repository"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

type MockUserRepository struct {
	createUserFunc        func(username, email, password string) (*repository.User, error)
	getUserByUsernameFunc func(username string) (*repository.User, error)
	getUserByEmailFunc    func(email string) (*repository.User, error)
}

var _ UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) CreateUser(username, email, password string) (*repository.User, error) {
	return m.createUserFunc(username, email, password)
}

func (m *MockUserRepository) GetUserByUsername(username string) (*repository.User, error) {
	return m.getUserByUsernameFunc(username)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*repository.User, error) {
	return m.getUserByEmailFunc(email)
}

func Test_userService_Register(t *testing.T) {
	const (
		jwtSecret     = "test-secret"
		jwtExpiration = time.Hour * 24
	)

	tests := []struct {
		name           string
		username       string
		email          string
		password       string
		mockCreateUser func(username, email, password string) (*repository.User, error)
		expectedError  error
		validateFunc   func(*User)
	}{
		{
			name:     "Valid registration",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			mockCreateUser: func(username, email, password string) (*repository.User, error) {
				if username != "testuser" || email != "test@example.com" {
					t.Errorf("Expected CreateUser(%q, %q, _), got CreateUser(%q, %q, _)", "testuser", "test@example.com", username, email)
				}

				// Verify password is hashed
				if password == "password" {
					t.Errorf("Password should be hashed, got plain text")
				}

				return &repository.User{
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
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}
				if !token.Valid {
					t.Errorf("Invalid token")
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
		{
			name:     "Username already taken",
			username: "existinguser",
			email:    "test@example.com",
			password: "password",
			mockCreateUser: func(username, email, password string) (*repository.User, error) {
				return nil, repository.ErrDuplicateUsername
			},
			expectedError: ErrUsernameTaken,
			validateFunc:  nil,
		},
		{
			name:     "Email already registered",
			username: "testuser",
			email:    "existing@example.com",
			password: "password",
			mockCreateUser: func(username, email, password string) (*repository.User, error) {
				return nil, repository.ErrDuplicateEmail
			},
			expectedError: ErrEmailTaken,
			validateFunc:  nil,
		},
		{
			name:     "Repository error",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			mockCreateUser: func(username, email, password string) (*repository.User, error) {
				return nil, repository.ErrInternal
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepository := &MockUserRepository{
				createUserFunc: tt.mockCreateUser,
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
