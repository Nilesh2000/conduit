package service

import (
	"context"
	"errors"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"
)

// Comment represents a comment on an article
type Comment struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Body      string    `json:"body"`
	Author    Profile   `json:"author"`
}

// CommentRepository is an interface for the comment repository
type CommentRepository interface {
	GetByID(ctx context.Context, commentID int64) (*repository.Comment, error)
	Create(ctx context.Context, userID, articleID int64, body string) (*repository.Comment, error)
	Delete(ctx context.Context, commentID int64) error
}

// commentService implements the CommentService interface
type commentService struct {
	commentRepository CommentRepository
	articleRepository ArticleRepository
}

// NewCommentService creates a new comment service
func NewCommentService(
	commentRepository CommentRepository,
	articleRepository ArticleRepository,
) *commentService {
	return &commentService{
		commentRepository: commentRepository,
		articleRepository: articleRepository,
	}
}

// CreateComment creates a new comment
func (s *commentService) CreateComment(
	ctx context.Context,
	userID int64,
	slug, body string,
) (*Comment, error) {
	article, err := s.articleRepository.GetBySlug(ctx, slug)
	if err != nil {
		switch {
		case errors.Is(err, ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	comment, err := s.commentRepository.Create(ctx, userID, article.ID, body)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrArticleNotFound):
			return nil, ErrArticleNotFound
		default:
			return nil, ErrInternalServer
		}
	}

	return &Comment{
		ID:        comment.ID,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Body:      comment.Body,
		Author: Profile{
			Username:  comment.Author.Username,
			Bio:       comment.Author.Bio,
			Image:     comment.Author.Image,
			Following: false,
		},
	}, nil
}

// DeleteComment deletes a comment
func (s *commentService) DeleteComment(
	ctx context.Context,
	userID int64,
	slug string,
	commentID int64,
) error {
	// Get the comment by ID
	comment, err := s.commentRepository.GetByID(ctx, commentID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrCommentNotFound):
			return ErrCommentNotFound
		default:
			return ErrInternalServer
		}
	}

	// Check if the comment is owned by the user
	if comment.Author.ID != userID {
		return ErrCommentNotAuthorized
	}

	// Delete the comment
	err = s.commentRepository.Delete(ctx, commentID)
	if err != nil {
		return ErrInternalServer
	}

	return nil
}
