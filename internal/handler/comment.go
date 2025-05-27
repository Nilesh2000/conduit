package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
)

type NewComment struct {
	Comment struct {
		Body string `json:"body" validate:"required"`
	} `json:"comment" validate:"required"`
}

type CommentService interface {
	CreateComment(ctx context.Context, userID int64, slug, body string) (*service.Comment, error)
}

type commentHandler struct {
	commentService CommentService
}

func NewCommentHandler(commentService CommentService) *commentHandler {
	return &commentHandler{commentService: commentService}
}

func (h *commentHandler) CreateComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		// Get slug from request path
		slug := r.PathValue("slug")

		// Get req body from request body
		var req NewComment
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.RespondWithError(w, http.StatusBadRequest, []string{"Invalid request body"})
			return
		}

		// Call service to create comment
		comment, err := h.commentService.CreateComment(r.Context(), userID, slug, req.Comment.Body)
		if err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
			return
		}

		// Respond with created comment
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(comment); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
			return
		}
	}
}
