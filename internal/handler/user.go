package handler

import (
	"conduit/internal/middleware"
	"conduit/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// GenericErrorModel represents the API error response body
type GenericErrorModel struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	User struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	} `json:"user"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	User struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Bio      string `json:"bio"`
		Image    string `json:"image"`
	} `json:"user"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	User struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	} `json:"user"`
}

// UserResponse represents the response body for user operations
type UserResponse struct {
	User service.User `json:"user"`
}

// UserService defines the interface for user service operations
type UserService interface {
	Register(username, email, password string) (*service.User, error)
	Login(email, password string) (*service.User, error)
	GetCurrentUser(userID int64) (*service.User, error)
	UpdateUser(userID int64, username, email, password, bio, image string) (*service.User, error)
}

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService UserService
	Validate    *validator.Validate
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		Validate:    validator.New(),
	}
}

// Register returns a handler function for user registration
func (h *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse request body
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Invalid request body"})
			return
		}

		// Validate request body
		if err := h.Validate.Struct(req); err != nil {
			errors := h.translateValidationErrors(err)
			h.respondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to register user
		user, err := h.userService.Register(req.User.Username, req.User.Email, req.User.Password)

		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUsernameTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Username already taken"})
			case errors.Is(err, service.ErrEmailTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Email already registered"})
			default:
				h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
			}
			return
		}

		// Respond with created user
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		})
	}
}

// Login returns a handler function for user login
func (h *UserHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse request body
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Invalid request body"})
			return
		}

		// Validate request body
		if err := h.Validate.Struct(req); err != nil {
			errors := h.translateValidationErrors(err)
			h.respondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to login user
		user, err := h.userService.Login(req.User.Email, req.User.Password)

		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserNotFound):
				h.respondWithError(w, http.StatusUnauthorized, []string{"Invalid credentials"})
			default:
				h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
			}
			return
		}

		// Respond with user data
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		})
	}
}

// GetCurrentUser returns a handler function for getting the current user
func (h *UserHandler) GetCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			h.respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Call service to get current user
		user, err := h.userService.GetCurrentUser(userID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				h.respondWithError(w, http.StatusNotFound, []string{"User not found"})
			default:
				h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
			}
			return
		}

		// Respond with user data
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		})
	}
}

// UpdateCurrentUser returns a handler function for updating the current user
func (h *UserHandler) UpdateCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			h.respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Parse request body
		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Invalid request body"})
			return
		}

		// Validate request body
		if err := h.Validate.Struct(req); err != nil {
			errors := h.translateValidationErrors(err)
			h.respondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to update user
		user, err := h.userService.UpdateUser(userID, req.User.Username, req.User.Email, req.User.Password, req.User.Bio, req.User.Image)

		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUsernameTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Username already taken"})
			case errors.Is(err, service.ErrEmailTaken):
				h.respondWithError(w, http.StatusUnprocessableEntity, []string{"Email already registered"})
			case errors.Is(err, service.ErrUserNotFound):
				h.respondWithError(w, http.StatusNotFound, []string{"User not found"})
			default:
				h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
			}
			return
		}

		// Respond with updated user
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		})
	}
}

// translateValidationErrors translates validation errors into a list of error messages
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

// respondWithError sends an error response with the given status code and errors
func (h *UserHandler) respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	json.NewEncoder(w).Encode(response)
}
