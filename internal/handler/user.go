package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"conduit/internal/middleware"
	"conduit/internal/response"
	"conduit/internal/service"
	"conduit/internal/validation"

	"github.com/go-playground/validator/v10"
)

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
		Username *string `json:"username" validate:"omitempty"`
		Email    *string `json:"email" validate:"omitempty,email"`
		Password *string `json:"password" validate:"omitempty,min=8"`
		Bio      *string `json:"bio" validate:"omitempty"`
		Image    *string `json:"image" validate:"omitempty"`
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
	Register(ctx context.Context, username, email, password string) (*service.User, error)
	Login(ctx context.Context, email, password string) (*service.User, error)
	GetCurrentUser(ctx context.Context, userID int64) (*service.User, error)
	UpdateUser(
		ctx context.Context,
		userID int64,
		username, email, password, bio, image *string,
	) (*service.User, error)
}

// userHandler handles user-related HTTP requests
type userHandler struct {
	userService UserService
	validate    *validator.Validate
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService UserService) *userHandler {
	return &userHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// Register returns a handler function for user registration
func (h *userHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse request body
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.RespondWithError(
				w,
				http.StatusUnprocessableEntity,
				[]string{"Invalid request body"},
			)
			return
		}

		// Validate request body
		if err := h.validate.Struct(req); err != nil {
			errors := validation.TranslateValidationErrors(err)
			response.RespondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to register user
		user, err := h.userService.Register(
			r.Context(),
			req.User.Username,
			req.User.Email,
			req.User.Password,
		)
		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUsernameTaken):
				response.RespondWithError(
					w,
					http.StatusUnprocessableEntity,
					[]string{"Username already taken"},
				)
			case errors.Is(err, service.ErrEmailTaken):
				response.RespondWithError(
					w,
					http.StatusUnprocessableEntity,
					[]string{"Email already registered"},
				)
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with created user
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// Login returns a handler function for user login
func (h *userHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse request body
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.RespondWithError(
				w,
				http.StatusUnprocessableEntity,
				[]string{"Invalid request body"},
			)
			return
		}

		// Validate request body
		if err := h.validate.Struct(req); err != nil {
			errors := validation.TranslateValidationErrors(err)
			response.RespondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to login user
		user, err := h.userService.Login(r.Context(), req.User.Email, req.User.Password)
		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(
					w,
					http.StatusUnauthorized,
					[]string{"Invalid credentials"},
				)
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with user data
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// GetCurrentUser returns a handler function for getting the current user
func (h *userHandler) GetCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get token from context
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Token ")

		// Call service to get current user
		user, err := h.userService.GetCurrentUser(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Add token to user
		user.Token = token

		// Respond with user data
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// UpdateCurrentUser returns a handler function for updating the current user
func (h *userHandler) UpdateCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get token from context
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Token ")

		// Parse request body
		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.RespondWithError(
				w,
				http.StatusUnprocessableEntity,
				[]string{"Invalid request body"},
			)
			return
		}

		// Validate request body
		if err := h.validate.Struct(req); err != nil {
			errors := validation.TranslateValidationErrors(err)
			response.RespondWithError(w, http.StatusUnprocessableEntity, errors)
			return
		}

		// Call service to update user
		user, err := h.userService.UpdateUser(
			r.Context(),
			userID,
			req.User.Username,
			req.User.Email,
			req.User.Password,
			req.User.Bio,
			req.User.Image,
		)
		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUsernameTaken):
				response.RespondWithError(
					w,
					http.StatusUnprocessableEntity,
					[]string{"Username already taken"},
				)
			case errors.Is(err, service.ErrEmailTaken):
				response.RespondWithError(
					w,
					http.StatusUnprocessableEntity,
					[]string{"Email already registered"},
				)
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Add token to user
		user.Token = token

		// Respond with updated user
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(UserResponse{
			User: *user,
		}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}
