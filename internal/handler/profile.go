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
)

// ProfileResponse is the response body for a profile
type ProfileResponse struct {
	Profile service.Profile `json:"profile"`
}

// ProfileService is an interface for the profile service
type ProfileService interface {
	GetProfile(ctx context.Context, username string, currentUserID int64) (*service.Profile, error)
	FollowUser(
		ctx context.Context,
		followerID int64,
		followingName string,
	) (*service.Profile, error)
	UnfollowUser(
		ctx context.Context,
		followerID int64,
		followingName string,
	) (*service.Profile, error)
}

// profileHandler is a handler for profile requests
type profileHandler struct {
	profileService ProfileService
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(profileService ProfileService) *profileHandler {
	return &profileHandler{
		profileService: profileService,
	}
}

// GetProfile gets a profile by username
func (h *profileHandler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get username from context
		username := r.URL.Path[len("/api/profiles/"):]

		// Get current user ID from context
		currentUserID := int64(0)
		if id, ok := middleware.GetUserIDFromContext(r.Context()); ok {
			currentUserID = id
		}

		// Get profile
		profile, err := h.profileService.GetProfile(r.Context(), username, currentUserID)
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

		// Respond with profile
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ProfileResponse{Profile: *profile}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// Follow follows a user
func (h *profileHandler) Follow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		followerID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get username from URL path
		pathParams := strings.Split(r.URL.Path, "/")
		if len(pathParams) < 4 {
			response.RespondWithError(w, http.StatusBadRequest, []string{"User not found"})
			return
		}
		// Get second to last path parameter (username in /profiles/:username/follow)
		username := pathParams[len(pathParams)-2]

		// Call service to follow user
		profile, err := h.profileService.FollowUser(r.Context(), followerID, username)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrCannotFollowSelf):
				response.RespondWithError(
					w,
					http.StatusBadRequest,
					[]string{"Cannot follow yourself"},
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

		// Respond with updated profile
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ProfileResponse{Profile: *profile}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// Unfollow unfollows a user
func (h *profileHandler) Unfollow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		followerID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get username from URL path
		pathParams := strings.Split(r.URL.Path, "/")
		if len(pathParams) < 4 {
			response.RespondWithError(w, http.StatusBadRequest, []string{"User not found"})
			return
		}

		// Get second to last path parameter (username in /profiles/:username/follow)
		username := pathParams[len(pathParams)-2]

		// Call service to unfollow user
		profile, err := h.profileService.UnfollowUser(r.Context(), followerID, username)
		// Handle errors
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrCannotFollowSelf):
				response.RespondWithError(
					w,
					http.StatusBadRequest,
					[]string{"Cannot unfollow yourself"},
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

		// Respond with updated profile
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(ProfileResponse{Profile: *profile}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}
