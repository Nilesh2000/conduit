package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"conduit/internal/repository"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	createFunc      func(username, email, password string) (*repository.User, error)
	findByEmailFunc func(email string) (*repository.User, error)
	findByIDFunc    func(id int64) (*repository.User, error)
	updateFunc      func(userID int64, username, email, password, bio, image *string) (*repository.User, error)
}

var _ UserRepository = (*MockUserRepository)(nil)

// Create creates a new user in the repository
func (m *MockUserRepository) Create(username, email, password string) (*repository.User, error) {
	return m.createFunc(username, email, password)
}

// FindByEmail finds a user by email in the repository
func (m *MockUserRepository) FindByEmail(email string) (*repository.User, error) {
	return m.findByEmailFunc(email)
}

// FindByID finds a user by ID in the repository
func (m *MockUserRepository) FindByID(id int64) (*repository.User, error) {
	return m.findByIDFunc(id)
}

// Update updates a user in the repository
func (m *MockUserRepository) Update(userID int64, username, email, password, bio, image *string) (*repository.User, error) {
	return m.updateFunc(userID, username, email, password, bio, image)
}

// Test_userService_Register tests the Register method of the userService
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
			password: "password123",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					createFunc: func(username, email, password string) (*repository.User, error) {
						if username != "testuser" || email != "test@example.com" {
							t.Errorf("Expected Create(%q, %q, _), got Create(%q, %q, _)", "testuser", "test@example.com", username, email)
						}

						err := bcrypt.CompareHashAndPassword([]byte(password), []byte("password123"))
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
				token, err := jwt.ParseWithClaims(u.Token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}
				if !token.Valid {
					t.Errorf("Token is not valid")
				}

				// Verify claims
				if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
					expectedSubject := "1"
					if claims.Subject != expectedSubject {
						t.Errorf("Expected token subject to be %q, got %q", expectedSubject, claims.Subject)
					}

					if claims.Issuer != "conduit-api" {
						t.Errorf("Expected token issuer to be 'conduit-api', got %q", claims.Issuer)
					}
				} else {
					t.Errorf("Failed to parse token claims as StandardClaims")
				}
			},
		},
		{
			name:     "Username already taken",
			username: "existinguser",
			email:    "test@example.com",
			password: "password123",
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
			password: "password123",
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
			password: "password123",
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
			// Setup mock repository
			mockUserRepository := tt.setupMock()

			// Create service with mock repository
			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			// Call Register
			user, err := userService.Register(tt.username, tt.email, tt.password)

			// Validate error
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Validate user if expected
			if err == nil && tt.validateFunc != nil {
				tt.validateFunc(t, user)
			}
		})
	}
}

// Test_userService_Login tests the Login method of the userService
func Test_userService_Login(t *testing.T) {
	const (
		jwtSecret     = "test-secret"
		jwtExpiration = time.Hour * 24
	)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMock     func() *MockUserRepository
		expectedError error
		validateFunc  func(t *testing.T, u *User)
	}{
		{
			name:     "Valid login",
			email:    "test@example.com",
			password: "password123",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByEmailFunc: func(email string) (*repository.User, error) {
						if email != "test@example.com" {
							t.Errorf("Expected FindByEmail(%q), got FindByEmail(%q)", "test@example.com", email)
						}

						hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
						if err != nil {
							t.Errorf("Failed to hash password: %v", err)
						}

						return &repository.User{
							ID:       1,
							Username: "testuser",
							Email:    "test@example.com",
							Password: string(hashedPassword),
							Bio:      "I'm a test user",
							Image:    "https://example.com/image.jpg",
						}, nil
					},
				}
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, u *User) {
				if u.Email != "test@example.com" {
					t.Errorf("Expected Email 'test@example.com', got %q", u.Email)
				}
				if u.Token == "" {
					t.Errorf("Expected non-empty token")
				}
				if u.Username != "testuser" {
					t.Errorf("Expected Username 'testuser', got %q", u.Username)
				}
				if u.Bio != "I'm a test user" {
					t.Errorf("Expected bio 'I'm a test user', got %q", u.Bio)
				}
				if u.Image != "https://example.com/image.jpg" {
					t.Errorf("Expected image 'https://example.com/image.jpg', got %q", u.Image)
				}

				// Verify token
				token, err := jwt.ParseWithClaims(u.Token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(jwtSecret), nil
				})
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}
				if !token.Valid {
					t.Errorf("Token is not valid")
				}

				// Verify claims
				if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
					expectedSubject := "1"
					if claims.Subject != expectedSubject {
						t.Errorf("Expected token subject to be %q, got %q", expectedSubject, claims.Subject)
					}

					if claims.Issuer != "conduit-api" {
						t.Errorf("Expected token issuer to be 'conduit-api', got %q", claims.Issuer)
					}
				} else {
					t.Errorf("Failed to parse token claims as StandardClaims")
				}
			},
		},
		{
			name:     "User not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByEmailFunc: func(email string) (*repository.User, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedError: ErrUserNotFound,
			validateFunc:  nil,
		},
		{
			name:     "Invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByEmailFunc: func(email string) (*repository.User, error) {
						hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
						if err != nil {
							t.Errorf("Failed to hash password: %v", err)
						}

						return &repository.User{
							ID:       1,
							Username: "testuser",
							Email:    "test@example.com",
							Password: string(hashedPassword),
							Bio:      "I'm a test user",
							Image:    "https://example.com/image.jpg",
						}, nil
					},
				}
			},
			expectedError: ErrInvalidCredentials,
			validateFunc:  nil,
		},
		{
			name:     "Repository error",
			email:    "test@example.com",
			password: "password123",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByEmailFunc: func(email string) (*repository.User, error) {
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
			// Setup mock repository
			mockUserRepository := tt.setupMock()

			// Create service with mock repository
			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			// Call Login
			user, err := userService.Login(tt.email, tt.password)

			// Validate error
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Validate user if expected
			if err == nil && tt.validateFunc != nil {
				tt.validateFunc(t, user)
			}
		})
	}
}
