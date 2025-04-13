package handler

import (
	"conduit/internal/middleware"
	"conduit/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// ProfileResponse is the response body for a profile
type ProfileResponse struct {
	Profile service.Profile `json:"profile"`
}

// ProfileService is an interface for the profile service
type ProfileService interface {
	GetProfile(username string, currentUserID int64) (*service.Profile, error)
}

// ProfileHandler is a handler for profile requests
type ProfileHandler struct {
	profileService ProfileService
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(profileService ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// GetProfile gets a profile by username
func (h *ProfileHandler) GetProfile() http.HandlerFunc {
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
				h.respondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
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

// respondWithError sends an error response with the given status code and errors
func (h *ProfileHandler) respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
