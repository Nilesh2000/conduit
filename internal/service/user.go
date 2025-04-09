package service

import (
	"conduit/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string
	Token    string
	Username string
	Bio      string
	Image    string
}

type UserRepository interface {
	Create(username, email, password string) (*repository.User, error)
	FindByEmail(email string) (*repository.User, error)
}

type userService struct {
	userRepository UserRepository
	jwtSecret      []byte
	jwtExpiration  time.Duration
}

func NewUserService(userRepository UserRepository, jwtSecret string, jwtExpiration time.Duration) *userService {
	return &userService{
		userRepository: userRepository,
		jwtSecret:      []byte(jwtSecret),
		jwtExpiration:  jwtExpiration,
	}
}

func (s *userService) Register(username, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternalServer
	}

	repoUser, err := s.userRepository.Create(username, email, string(hashedPassword))
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

	token, err := s.generateToken(repoUser.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	return &User{
		Email:    repoUser.Email,
		Token:    token,
		Username: repoUser.Username,
		Bio:      repoUser.Bio,
		Image:    repoUser.Image,
	}, nil
}

func (s *userService) Login(email, password string) (*User, error) {
	repoUser, err := s.userRepository.FindByEmail(email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(repoUser.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(repoUser.ID)
	if err != nil {
		return nil, ErrInternalServer
	}

	return &User{
		Email:    repoUser.Email,
		Token:    token,
		Username: repoUser.Username,
		Bio:      repoUser.Bio,
		Image:    repoUser.Image,
	}, nil
}

func (s *userService) generateToken(userID int64) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.jwtExpiration)

	claims := jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Id:        uuid.New().String(),
		IssuedAt:  now.Unix(),
		Issuer:    "conduit-api",
		NotBefore: now.Unix(),
		Subject:   fmt.Sprintf("%d", userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
