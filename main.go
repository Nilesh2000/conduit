package main

import (
	"encoding/json"
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

type UserService interface {
	Register(username, email, password string) (*User, error)
}

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

func (h *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Invalid request body"})
			return
		}

		if err := h.Validate.Struct(req); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(ErrorResponse{
				Errors: struct {
					Body []string
				}{
					Body: []string{"Invalid request body"},
				},
			})
			return
		}

		user, _ := h.userService.Register(req.User.Username, req.User.Email, req.User.Password)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		})
	}
}
