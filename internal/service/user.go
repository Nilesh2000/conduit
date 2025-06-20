package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	Create(ctx context.Context, username, email, password string) (*repository.User, error)
	FindByEmail(ctx context.Context, email string) (*repository.User, error)
	FindByID(ctx context.Context, id int64) (*repository.User, error)
	Update(
		ctx context.Context,
		userID int64,
		username, email, passwordHash, bio, image *string,
	) (*repository.User, error)
}

// userService implements the UserService interface
type userService struct {
	userRepository UserRepository
	jwtSecret      []byte
	jwtExpiration  time.Duration
}

// NewUserService creates a new user service
func NewUserService(
	userRepository UserRepository,
	jwtSecret string,
	jwtExpiration time.Duration,
) *userService {
	return &userService{
		userRepository: userRepository,
		jwtSecret:      []byte(jwtSecret),
		jwtExpiration:  jwtExpiration,
	}
}

// Register creates a new user in the system
func (s *userService) Register(
	ctx context.Context,
	username, email, password string,
) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Create the user in the repository
	user, err := s.userRepository.Create(ctx, username, email, string(hashedPassword))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateUsername):
			return nil, ErrUsernameTaken
		case errors.Is(err, repository.ErrDuplicateEmail):
			return nil, ErrEmailTaken
		default:
			return nil, ErrInternalServer
		}
	}

	// Generate a JWT token for the user
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Return user data
	return &User{
		Email:    user.Email,
		Token:    token,
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}, nil
}

// Login authenticates a user with email and password
func (s *userService) Login(ctx context.Context, email, password string) (*User, error) {
	// Find the user by email
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	// Return user data
	return &User{
		Email:    user.Email,
		Token:    token,
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}, nil
}

// GetCurrentUser gets the current user in the system
func (s *userService) GetCurrentUser(ctx context.Context, userID int64) (*User, error) {
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	return &User{
		Email:    user.Email,
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}, nil
}

// UpdateUser updates a user in the system
func (s *userService) UpdateUser(
	ctx context.Context,
	userID int64,
	username, email, password, bio, image *string,
) (*User, error) {
	// Hash the password if provided
	var hashedPassword *string
	if password != nil {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, ErrInternalServer
		}
		h := string(hashedBytes)
		hashedPassword = &h
	}

	user, err := s.userRepository.Update(ctx, userID, username, email, hashedPassword, bio, image)

	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		return nil, ErrUserNotFound
	case errors.Is(err, repository.ErrDuplicateUsername):
		return nil, ErrUsernameTaken
	case errors.Is(err, repository.ErrDuplicateEmail):
		return nil, ErrEmailTaken
	case errors.Is(err, repository.ErrInternal):
		return nil, ErrInternalServer
	}

	return &User{
		Email:    user.Email,
		Username: user.Username,
		Bio:      user.Bio,
		Image:    user.Image,
	}, nil
}

// generateToken generates a JWT token for a user
func (s *userService) generateToken(userID int64) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.jwtExpiration)

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		ID:        uuid.New().String(),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "conduit-api",
		NotBefore: jwt.NewNumericDate(now),
		Subject:   fmt.Sprintf("%d", userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
