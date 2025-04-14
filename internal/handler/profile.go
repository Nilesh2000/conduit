package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"conduit/internal/middleware"
	"conduit/internal/service"
)

// ProfileResponse is the response body for a profile
type ProfileResponse struct {
	Profile service.Profile `json:"profile"`
}

// ProfileService is an interface for the profile service
type ProfileService interface {
	GetProfile(username string, currentUserID int64) (*service.Profile, error)
	FollowUser(followerID int64, followingName string) (*service.Profile, error)
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
		profile, err := h.profileService.GetProfile(username, currentUserID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				h.respondWithError(w, http.StatusNotFound, []string{"User not found"})
			default:
				h.respondWithError(
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
			h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
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
			h.respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get username from URL path
		pathParams := strings.Split(r.URL.Path, "/")
		if len(pathParams) < 4 {
			h.respondWithError(w, http.StatusBadRequest, []string{"User not found"})
			return
		}
		// Get second to last path parameter (username in /profiles/:username/follow)
		username := pathParams[len(pathParams)-2]

		// Call service to follow user
		profile, err := h.profileService.FollowUser(followerID, username)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				h.respondWithError(w, http.StatusNotFound, []string{"User not found"})
			case errors.Is(err, service.ErrCannotFollowSelf):
				h.respondWithError(w, http.StatusForbidden, []string{"Cannot follow yourself"})
			default:
				h.respondWithError(
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
			h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
		}
	}
}

// respondWithError sends an error response with the given status code and errors
func (h *profileHandler) respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
