package service

import "context"

// TagRepository defines the interface for tag repository operations
type TagRepository interface {
	Get(ctx context.Context) ([]string, error)
}

// tagService implements the TagService interface
type tagService struct {
	tagRepository TagRepository
}

// NewTagService creates a new tag service
func NewTagService(tagRepository TagRepository) *tagService {
	return &tagService{tagRepository: tagRepository}
}

// GetTags gets all tags
func (s *tagService) GetTags(ctx context.Context) ([]string, error) {
	return s.tagRepository.Get(ctx)
}
