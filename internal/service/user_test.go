package service

import (
	"conduit/internal/repository"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	createFunc func(username, email, password string) (*repository.User, error)
}

var _ UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) Create(username, email, password string) (*repository.User, error) {
	return m.createFunc(username, email, password)
}

func Test_userService_Register(t *testing.T) {
	const (
		jwtSecret     = "test-secret"
		jwtExpiration = time.Hour * 24
	)

	tests := []struct {
		name          string
		username      string
		email         string
		password      string
		setupMock     func() *MockUserRepository
		expectedError error
		validateFunc  func(t *testing.T, u *User)
	}{
		{
			name:     "Valid registration",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					createFunc: func(username, email, password string) (*repository.User, error) {
						if username != "testuser" || email != "test@example.com" {
							t.Errorf("Expected Create(%q, %q, _), got Create(%q, %q, _)", "testuser", "test@example.com", username, email)
						}

						err := bcrypt.CompareHashAndPassword([]byte(password), []byte("password"))
						if err != nil {
							t.Errorf("Password not properly hashed: %v", err)
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
				}
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, u *User) {
				if u.Username != "testuser" {
					t.Errorf("Expected Username 'testuser', got %q", u.Username)
				}
				if u.Email != "test@example.com" {
					t.Errorf("Expected Email 'test@example.com', got %q", u.Email)
				}
				if u.Token == "" {
					t.Errorf("Expected non-empty token")
				}
				if u.Bio != "" {
					t.Errorf("Expected empty bio, got %q", u.Bio)
				}
				if u.Image != "" {
					t.Errorf("Expected empty image, got %q", u.Image)
				}

				// Verify token
				token, err := jwt.Parse(u.Token, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}
				if !token.Valid {
					t.Errorf("Token is not valid")
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
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					createFunc: func(username, email, password string) (*repository.User, error) {
						return nil, repository.ErrDuplicateUsername
					},
				}
			},
			expectedError: ErrUsernameTaken,
			validateFunc:  nil,
		},
		{
			name:     "Email already taken",
			username: "testuser",
			email:    "existing@example.com",
			password: "password",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					createFunc: func(username, email, password string) (*repository.User, error) {
						return nil, repository.ErrDuplicateEmail
					},
				}
			},
			expectedError: ErrEmailTaken,
			validateFunc:  nil,
		},
		{
			name:     "Password hashing error",
			username: "testuser",
			email:    "test@example.com",
			password: strings.Repeat("a", 73), // 73 bytes will exceed bcrypt's limit
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					createFunc: func(username, email, password string) (*repository.User, error) {
						t.Errorf("Create should not be called when password hashing fails")
						return nil, nil
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
		{
			name:     "Repository error",
			username: "testuser",
			email:    "test@example.com",
			password: "password",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					createFunc: func(username, email, password string) (*repository.User, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepository := tt.setupMock()

			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			user, err := userService.Register(tt.username, tt.email, tt.password)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			if err == nil && tt.validateFunc != nil {
				tt.validateFunc(t, user)
			}
		})
	}
}
