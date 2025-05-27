package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	createFunc      func(ctx context.Context, username, email, password string) (*repository.User, error)
	findByEmailFunc func(ctx context.Context, email string) (*repository.User, error)
	findByIDFunc    func(ctx context.Context, id int64) (*repository.User, error)
	updateFunc      func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error)
}

var _ UserRepository = (*MockUserRepository)(nil)

// Create creates a new user in the repository
func (m *MockUserRepository) Create(
	ctx context.Context,
	username, email, password string,
) (*repository.User, error) {
	return m.createFunc(ctx, username, email, password)
}

// FindByEmail finds a user by email in the repository
func (m *MockUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*repository.User, error) {
	return m.findByEmailFunc(ctx, email)
}

// FindByID finds a user by ID in the repository
func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*repository.User, error) {
	return m.findByIDFunc(ctx, id)
}

// Update updates a user in the repository
func (m *MockUserRepository) Update(
	ctx context.Context,
	userID int64,
	username, email, password, bio, image *string,
) (*repository.User, error) {
	return m.updateFunc(ctx, userID, username, email, password, bio, image)
}

// Test_userService_Register tests the Register method of the userService
func Test_userService_Register(t *testing.T) {
	t.Parallel()

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
					createFunc: func(ctx context.Context, username, email, password string) (*repository.User, error) {
						if username != "testuser" || email != "test@example.com" {
							t.Errorf(
								"Expected Create(%q, %q, _), got Create(%q, %q, _)",
								"testuser",
								"test@example.com",
								username,
								email,
							)
						}

						err := bcrypt.CompareHashAndPassword(
							[]byte(password),
							[]byte("password123"),
						)
						if err != nil {
							t.Errorf("Password not properly hashed: %v", err)
						}

						return &repository.User{
							ID:           1,
							Username:     username,
							Email:        email,
							PasswordHash: password,
							Bio:          "",
							Image:        "",
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
				token, err := jwt.ParseWithClaims(
					u.Token,
					&jwt.RegisteredClaims{},
					func(token *jwt.Token) (any, error) {
						return []byte(jwtSecret), nil
					},
				)
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}
				if !token.Valid {
					t.Errorf("Token is not valid")
				}

				// Verify claims
				if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
					expectedSubject := "1"
					if claims.Subject != expectedSubject {
						t.Errorf(
							"Expected token subject to be %q, got %q",
							expectedSubject,
							claims.Subject,
						)
					}

					if claims.Issuer != "conduit-api" {
						t.Errorf("Expected token issuer to be 'conduit-api', got %q", claims.Issuer)
					}
				} else {
					t.Errorf("Failed to parse token claims as RegisteredClaims")
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
					createFunc: func(ctx context.Context, username, email, password string) (*repository.User, error) {
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
					createFunc: func(ctx context.Context, username, email, password string) (*repository.User, error) {
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
					createFunc: func(ctx context.Context, username, email, password string) (*repository.User, error) {
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
					createFunc: func(ctx context.Context, username, email, password string) (*repository.User, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockUserRepository := tt.setupMock()

			// Create service with mock repository
			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			// Create context
			ctx := context.Background()

			// Call Register
			user, err := userService.Register(ctx, tt.username, tt.email, tt.password)

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
	t.Parallel()

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
					findByEmailFunc: func(ctx context.Context, email string) (*repository.User, error) {
						if email != "test@example.com" {
							t.Errorf(
								"Expected FindByEmail(%q), got FindByEmail(%q)",
								"test@example.com",
								email,
							)
						}

						hashedPassword, err := bcrypt.GenerateFromPassword(
							[]byte("password123"),
							bcrypt.DefaultCost,
						)
						if err != nil {
							t.Errorf("Failed to hash password: %v", err)
						}

						return &repository.User{
							ID:           1,
							Username:     "testuser",
							Email:        "test@example.com",
							PasswordHash: string(hashedPassword),
							Bio:          "I'm a test user",
							Image:        "https://example.com/image.jpg",
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
				token, err := jwt.ParseWithClaims(
					u.Token,
					&jwt.RegisteredClaims{},
					func(token *jwt.Token) (any, error) {
						return []byte(jwtSecret), nil
					},
				)
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}
				if !token.Valid {
					t.Errorf("Token is not valid")
				}

				// Verify claims
				if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
					expectedSubject := "1"
					if claims.Subject != expectedSubject {
						t.Errorf(
							"Expected token subject to be %q, got %q",
							expectedSubject,
							claims.Subject,
						)
					}

					if claims.Issuer != "conduit-api" {
						t.Errorf("Expected token issuer to be 'conduit-api', got %q", claims.Issuer)
					}
				} else {
					t.Errorf("Failed to parse token claims as RegisteredClaims")
				}
			},
		},
		{
			name:     "User not found",
			email:    "nonexistent@example.com",
			password: "password123",
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByEmailFunc: func(ctx context.Context, email string) (*repository.User, error) {
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
					findByEmailFunc: func(ctx context.Context, email string) (*repository.User, error) {
						hashedPassword, err := bcrypt.GenerateFromPassword(
							[]byte("password123"),
							bcrypt.DefaultCost,
						)
						if err != nil {
							t.Errorf("Failed to hash password: %v", err)
						}

						return &repository.User{
							ID:           1,
							Username:     "testuser",
							Email:        "test@example.com",
							PasswordHash: string(hashedPassword),
							Bio:          "I'm a test user",
							Image:        "https://example.com/image.jpg",
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
					findByEmailFunc: func(ctx context.Context, email string) (*repository.User, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockUserRepository := tt.setupMock()

			// Create service with mock repository
			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			// Create context
			ctx := context.Background()

			// Call Login
			user, err := userService.Login(ctx, tt.email, tt.password)

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

// Test_userService_GetCurrentUser tests the GetCurrentUser method of the userService
func Test_userService_GetCurrentUser(t *testing.T) {
	t.Parallel()

	const (
		jwtSecret     = "test-secret"
		jwtExpiration = time.Hour * 24
	)

	tests := []struct {
		name          string
		userID        int64
		setupMock     func() *MockUserRepository
		expectedError error
		validateFunc  func(t *testing.T, u *User)
	}{
		{
			name:   "Valid User ID",
			userID: 1,
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByIDFunc: func(ctx context.Context, id int64) (*repository.User, error) {
						if id != 1 {
							t.Errorf("Expected UserID 1, got %d", id)
						}

						return &repository.User{
							ID:           1,
							Username:     "testuser",
							Email:        "test@example.com",
							PasswordHash: "password123",
							Bio:          "I'm a test user",
							Image:        "https://example.com/image.jpg",
						}, nil
					},
				}
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, u *User) {
				if u.Email != "test@example.com" {
					t.Errorf("Expected Email 'test@example.com', got %q", u.Email)
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
				if u.Token != "" {
					t.Errorf("Expected token to be empty, got %q", u.Token)
				}
			},
		},
		{
			name:   "User not found",
			userID: 999,
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByIDFunc: func(ctx context.Context, id int64) (*repository.User, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedError: ErrUserNotFound,
			validateFunc:  nil,
		},
		{
			name:   "Repository error",
			userID: 1,
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					findByIDFunc: func(ctx context.Context, id int64) (*repository.User, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockUserRepository := tt.setupMock()

			// Create service with mock repository
			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			// Create context
			ctx := context.Background()

			// Call GetCurrentUser
			user, err := userService.GetCurrentUser(ctx, tt.userID)

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

// Test_userService_UpdateUser tests the UpdateUser method of the userService
func Test_userService_UpdateUser(t *testing.T) {
	t.Parallel()

	const (
		jwtSecret     = "test-secret"
		jwtExpiration = time.Hour * 24
	)

	// Helper functions to create pointers to strings
	strPtr := func(s string) *string {
		return &s
	}

	// Helper function to create a nil pointer to a string
	nilStrPtr := func() *string {
		return nil
	}

	tests := []struct {
		name          string
		userID        int64
		username      *string
		email         *string
		password      *string
		bio           *string
		image         *string
		setupMock     func() *MockUserRepository
		expectedError error
		validateFunc  func(t *testing.T, u *User)
	}{
		{
			name:     "Update all fields",
			userID:   1,
			username: strPtr("updateduser"),
			email:    strPtr("updated@example.com"),
			password: strPtr("newpassword123"),
			bio:      strPtr("Updated bio"),
			image:    strPtr("https://example.com/updated-image.jpg"),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						if userID != 1 {
							t.Errorf("Expected UserID 1, got %d", userID)
						}
						if *username != "updateduser" {
							t.Errorf("Expected Username 'updateduser', got %q", *username)
						}
						if *email != "updated@example.com" {
							t.Errorf("Expected Email 'updated@example.com', got %q", *email)
						}
						if !strings.HasPrefix(*password, "$2a$") {
							t.Errorf("Expected password to be bcrypt hashed, got %q", *password)
						}
						if *bio != "Updated bio" {
							t.Errorf("Expected Bio 'Updated bio', got %q", *bio)
						}
						if *image != "https://example.com/updated-image.jpg" {
							t.Errorf(
								"Expected Image 'https://example.com/updated-image.jpg', got %q",
								*image,
							)
						}

						return &repository.User{
							ID:           1,
							Username:     "updateduser",
							Email:        "updated@example.com",
							PasswordHash: "newpassword123",
							Bio:          "Updated bio",
							Image:        "https://example.com/updated-image.jpg",
						}, nil
					},
				}
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, u *User) {
				if u.Username != "updateduser" {
					t.Errorf("Expected Username 'updateduser', got %q", u.Username)
				}
				if u.Email != "updated@example.com" {
					t.Errorf("Expected Email 'updated@example.com', got %q", u.Email)
				}
				if u.Bio != "Updated bio" {
					t.Errorf("Expected bio 'Updated bio', got %q", u.Bio)
				}
				if u.Image != "https://example.com/updated-image.jpg" {
					t.Errorf(
						"Expected image 'https://example.com/updated-image.jpg', got %q",
						u.Image,
					)
				}

				// Token should not be set by UpdateUser
				if u.Token != "" {
					t.Errorf("Expected token to be empty, got %q", u.Token)
				}
			},
		},
		{
			name:     "Partial update - bio and image only",
			userID:   1,
			username: nilStrPtr(),
			email:    nilStrPtr(),
			password: nilStrPtr(),
			bio:      strPtr("Updated bio"),
			image:    strPtr("https://example.com/updated-image.jpg"),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						if userID != 1 {
							t.Errorf("Expected UserID 1, got %d", userID)
						}
						if username != nil {
							t.Errorf("Expected Username to be nil, got %q", *username)
						}
						if email != nil {
							t.Errorf("Expected Email to be nil, got %q", *email)
						}
						if password != nil {
							t.Errorf("Expected Password to be nil, got %q", *password)
						}
						if *bio != "Updated bio" {
							t.Errorf("Expected Bio 'Updated bio', got %q", *bio)
						}
						if *image != "https://example.com/updated-image.jpg" {
							t.Errorf(
								"Expected Image 'https://example.com/updated-image.jpg', got %q",
								*image,
							)
						}

						return &repository.User{
							ID:           1,
							Username:     "existinguser",
							Email:        "existing@example.com",
							PasswordHash: "existingpasswordhash",
							Bio:          *bio,
							Image:        *image,
						}, nil
					},
				}
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, u *User) {
				if u.Username != "existinguser" {
					t.Errorf("Expected Username 'existinguser', got %q", u.Username)
				}
				if u.Email != "existing@example.com" {
					t.Errorf("Expected Email 'existing@example.com', got %q", u.Email)
				}
				if u.Bio != "Updated bio" {
					t.Errorf("Expected Bio 'Updated bio', got %q", u.Bio)
				}
				if u.Image != "https://example.com/updated-image.jpg" {
					t.Errorf(
						"Expected Image 'https://example.com/updated-image.jpg', got %q",
						u.Image,
					)
				}
			},
		},
		{
			name:     "User not found",
			userID:   999,
			username: strPtr("updateduser"),
			email:    strPtr("updated@example.com"),
			password: strPtr("newpassword123"),
			bio:      strPtr("Updated bio"),
			image:    strPtr("https://example.com/updated-image.jpg"),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedError: ErrUserNotFound,
			validateFunc:  nil,
		},
		{
			name:     "Username already taken",
			userID:   1,
			username: strPtr("existinguser"),
			email:    strPtr("updated@example.com"),
			password: strPtr("newpassword123"),
			bio:      strPtr("Updated bio"),
			image:    strPtr("https://example.com/updated-image.jpg"),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						return nil, repository.ErrDuplicateUsername
					},
				}
			},
			expectedError: ErrUsernameTaken,
			validateFunc:  nil,
		},
		{
			name:     "Email already taken",
			userID:   1,
			username: strPtr("updateduser"),
			email:    strPtr("existing@example.com"),
			password: strPtr("newpassword123"),
			bio:      strPtr("Updated bio"),
			image:    strPtr("https://example.com/updated-image.jpg"),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						return nil, repository.ErrDuplicateEmail
					},
				}
			},
			expectedError: ErrEmailTaken,
			validateFunc:  nil,
		},
		{
			name:     "Repository error",
			userID:   1,
			username: strPtr("updateduser"),
			email:    strPtr("updated@example.com"),
			password: strPtr("newpassword123"),
			bio:      strPtr("Updated bio"),
			image:    strPtr("https://example.com/updated-image.jpg"),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						return nil, repository.ErrInternal
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
		{
			name:     "Password hashing error",
			userID:   1,
			username: strPtr("updateduser"),
			email:    strPtr("updated@example.com"),
			password: strPtr(strings.Repeat("a", 73)),
			setupMock: func() *MockUserRepository {
				return &MockUserRepository{
					updateFunc: func(ctx context.Context, userID int64, username, email, password, bio, image *string) (*repository.User, error) {
						t.Errorf("Update should not be called when password hashing fails")
						return nil, nil
					},
				}
			},
			expectedError: ErrInternalServer,
			validateFunc:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockUserRepository := tt.setupMock()

			// Create service with mock repository
			userService := NewUserService(mockUserRepository, jwtSecret, jwtExpiration)

			// Create context
			ctx := context.Background()

			// Call UpdateUser
			user, err := userService.UpdateUser(
				ctx,
				tt.userID,
				tt.username,
				tt.email,
				tt.password,
				tt.bio,
				tt.image,
			)

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
