package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Nilesh2000/conduit/internal/response"
)

type TagResponse struct {
	Tags []string `json:"tags"`
}

type TagService interface {
	GetTags(ctx context.Context) ([]string, error)
}

type tagHandler struct {
	tagService TagService
}

func NewTagHandler(tagService TagService) *tagHandler {
	return &tagHandler{tagService: tagService}
}

func (h *tagHandler) GetTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get tags from service
		tags, err := h.tagService.GetTags(r.Context())
		if err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
			return
		}

		// Respond with tags
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(TagResponse{Tags: tags}); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, []string{"Internal server error"})
		}
	}
}
