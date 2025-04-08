package handler

import (
	"conduit/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type GenericErrorModel struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

type ErrorResponse = GenericErrorModel

type RegisterRequest struct {
	User struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	} `json:"user"`
}

type LoginRequest struct {
	User struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}

type UserResponse struct {
	User service.User `json:"user"`
}

type UserService interface {
	Register(username, email, password string) (*service.User, error)
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
			case errors.Is(err, service.ErrUsernameTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"username already taken"})
			case errors.Is(err, service.ErrEmailTaken):
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

func (h *UserHandler) respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := ErrorResponse{}
	response.Errors.Body = errors

	json.NewEncoder(w).Encode(response)
}
