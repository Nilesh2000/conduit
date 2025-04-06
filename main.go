package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type LoginUser struct {
	Email    string
	Password string
}

type NewUser struct {
	Username string
	Email    string
	Password string
}

type User struct {
	Email    string
	Token    string
	Username string
	Bio      string
	Image    string
}

type UpdateUser struct {
	Email    string
	Token    string
	Username string
	Bio      string
	Image    string
}

type Profile struct {
	Username  string
	Bio       string
	Image     string
	Following bool
}

type Article struct {
	Slug           string
	Title          string
	Description    string
	Body           string
	TagList        []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Favorited      bool
	FavoritesCount int
	Author         Profile
}

type NewArticle struct {
	Title       string
	Description string
	Body        string
	TagList     []string
}

type UpdateArticle struct {
	Title       string
	Description string
	Body        string
}

type Comment struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Body      string
	Author    Profile
}

type NewComment struct {
	Body string
}

type GenericErrorModel struct {
	Errors struct {
		Body []string
	}
}

type UserRepo struct {
	ID       int64
	Username string
	Email    string
	Password string
	Bio      string
	Image    string
}

type UserRepository interface {
	CreateUser(username, email, password string) (*UserRepo, error)
	GetUserByUsername(username string) (*UserRepo, error)
	GetUserByEmail(email string) (*UserRepo, error)
}

type UserService interface {
	Register(username, email, password string) (*User, error)
}

var (
	ErrUsernameTaken  = errors.New("username already taken")
	ErrEmailTaken     = errors.New("email already registered")
	ErrInternalServer = errors.New("internal server error")
)

type RegisterRequest struct {
	User struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	} `json:"user"`
}

type UserResponse struct {
	User User `json:"user"`
}

type ErrorResponse = GenericErrorModel

type userService struct {
	userRepository UserRepository
	jwtSecret      []byte
	jwtExpiration  time.Duration
}

func (s *userService) Register(username, email, password string) (*User, error) {
	return nil, nil
}

type UserHandler struct {
	userService UserService
	Validate    *validator.Validate
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		Validate:    validator.New(),
	}
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := ErrorResponse{}
	response.Errors.Body = errors

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) translateValidationErrors(err error) []string {
	var validationErrors []string

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			switch e.Tag() {
			case "required":
				validationErrors = append(validationErrors, fmt.Sprintf("%s is required", e.Field()))
			case "email":
				validationErrors = append(validationErrors, fmt.Sprintf("%s is not a valid email", e.Value()))
			case "min":
				validationErrors = append(validationErrors, fmt.Sprintf("%s must be at least %s characters long", e.Field(), e.Param()))
			default:
				validationErrors = append(validationErrors, fmt.Sprintf("%s is not valid", e.Field()))
			}
		}
	} else {
		validationErrors = append(validationErrors, "Invalid request body")
	}

	return validationErrors
}

func (h *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Invalid request body"})
			return
		}

		if err := h.Validate.Struct(req); err != nil {
			errors := h.translateValidationErrors(err)
			h.respondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		user, err := h.userService.Register(req.User.Username, req.User.Email, req.User.Password)
		if err != nil {
			switch {
			case errors.Is(err, ErrUsernameTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"username already taken"})
			case errors.Is(err, ErrEmailTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"email already registered"})
			default:
				h.respondWithError(w, http.StatusInternalServerError, []string{"internal server error"})
			}
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		})
	}
}
