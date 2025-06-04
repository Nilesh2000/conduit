package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/response"
	"github.com/Nilesh2000/conduit/internal/service"
)

// NewComment is the request for a new comment
type NewComment struct {
	Comment struct {
		Body string `json:"body" validate:"required"`
	} `json:"comment" validate:"required"`
}

// CommentResponse is the response for a comment
type CommentResponse struct {
	Comment service.Comment `json:"comment"`
}

type CommentsResponse struct {
	Comments []service.Comment `json:"comments"`
}

// CommentService is an interface for the comment service
type CommentService interface {
	GetComments(ctx context.Context, slug string, userID *int64) ([]service.Comment, error)
	CreateComment(ctx context.Context, userID int64, slug, body string) (*service.Comment, error)
	DeleteComment(ctx context.Context, userID int64, slug string, commentID int64) error
}

// commentHandler is a handler for comment-related requests
type commentHandler struct {
	commentService CommentService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentService CommentService) *commentHandler {
	return &commentHandler{commentService: commentService}
}

// GetComments is a handler for getting comments for an article
func (h *commentHandler) GetComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get current user ID from context (optional)
		var userID *int64
		if id, ok := middleware.GetUserIDFromContext(r.Context()); ok {
			userID = &id
		}

		slug := r.PathValue("slug")

		comments, err := h.commentService.GetComments(r.Context(), slug, userID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrArticleNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(CommentsResponse{Comments: comments}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
		}
	}
}

// CreateComment is a handler for creating a new comment
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
			switch {
			case errors.Is(err, service.ErrArticleNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Article not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
			return
		}

		// Respond with created comment
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(CommentResponse{Comment: *comment}); err != nil {
			response.RespondWithError(
				w,
				http.StatusInternalServerError,
				[]string{"Internal server error"},
			)
			return
		}
	}
}

func (h *commentHandler) DeleteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get user ID from context
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
			return
		}

		slug := r.PathValue("slug")
		commentID := r.PathValue("id")
		commentIDInt, err := strconv.ParseInt(commentID, 10, 64)
		if err != nil {
			response.RespondWithError(w, http.StatusBadRequest, []string{"Invalid comment ID"})
			return
		}

		err = h.commentService.DeleteComment(r.Context(), userID, slug, commentIDInt)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrCommentNotAuthorized):
				response.RespondWithError(
					w,
					http.StatusForbidden,
					[]string{"You are not the author of this comment"},
				)
			case errors.Is(err, service.ErrCommentNotFound):
				response.RespondWithError(w, http.StatusNotFound, []string{"Comment not found"})
			default:
				response.RespondWithError(
					w,
					http.StatusInternalServerError,
					[]string{"Internal server error"},
				)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
