package service

import (
	"conduit/internal/repository"
	"errors"
	"testing"
	"time"

	"github.com/gosimple/slug"
)

// MockArticleRepository is a mock implementation of the ArticleRepository interface
type MockArticleRepository struct {
	createFunc    func(userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error)
	getBySlugFunc func(slug string) (*repository.Article, error)
}

// Create is a mock implementation of the Create method
func (m *MockArticleRepository) Create(
	userID int64,
	articleSlug, title, description, body string,
	tagList []string,
) (*repository.Article, error) {
	return m.createFunc(userID, articleSlug, title, description, body, tagList)
}

// GetBySlug is a mock implementation of the GetBySlug method
func (m *MockArticleRepository) GetBySlug(slug string) (*repository.Article, error) {
	return m.getBySlugFunc(slug)
}

func Test_articleService_CreateArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      int64
		title       string
		description string
		body        string
		tagList     []string
		setupMock   func() *MockArticleRepository
		expectedErr error
		validate    func(*testing.T, *Article)
	}{
		{
			name:        "Successful creation",
			userID:      1,
			title:       "Test Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{"tag1", "tag2"},
			setupMock: func() *MockArticleRepository {
				mockRepo := &MockArticleRepository{
					createFunc: func(userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						expectedSlug := slug.Make("Test Article")
						if articleSlug != expectedSlug {
							t.Errorf("Expected slug %q, got %q", expectedSlug, articleSlug)
						}
						if title != "Test Article" {
							t.Errorf("Expected title %q, got %q", "Test Article", title)
						}
						if description != "Test Description" {
							t.Errorf("Expected description %q, got %q", "Test Description", description)
						}
						if body != "Test Body" {
							t.Errorf("Expected body %q, got %q", "Test Body", body)
						}
						if len(tagList) != 2 || tagList[0] != "tag1" || tagList[1] != "tag2" {
							t.Errorf("Expected tags %v, got %v", []string{"tag1", "tag2"}, tagList)
						}

						now := time.Now()
						return &repository.Article{
							ID:          1,
							Slug:        expectedSlug,
							Title:       title,
							Description: description,
							Body:        body,
							AuthorID:    1,
							Author: &repository.User{
								ID:       1,
								Username: "testuser",
								Bio:      "Test Bio",
								Image:    "https://example.com/image.jpg",
							},
							CreatedAt: now,
							UpdatedAt: now,
							TagList:   []string{"tag1", "tag2"},
						}, nil
					},
				}
				return mockRepo
			},
			expectedErr: nil,
			validate: func(t *testing.T, article *Article) {
				expectedSlug := slug.Make("Test Article")
				if article.Slug != expectedSlug {
					t.Errorf("Expected slug %q, got %q", expectedSlug, article.Slug)
				}
				if article.Title != "Test Article" {
					t.Errorf("Expected title %q, got %q", "Test Article", article.Title)
				}
				if article.Description != "Test Description" {
					t.Errorf("Expected description %q, got %q", "Test Description", article.Description)
				}
				if article.Body != "Test Body" {
					t.Errorf("Expected body %q, got %q", "Test Body", article.Body)
				}
				if len(article.TagList) != 2 || article.TagList[0] != "tag1" || article.TagList[1] != "tag2" {
					t.Errorf("Expected tags %v, got %v", []string{"tag1", "tag2"}, article.TagList)
				}
				if article.Favorited {
					t.Errorf("Expected favorited to be false, got true")
				}
				if article.FavoritesCount != 0 {
					t.Errorf("Expected favorites count to be 0, got %d", article.FavoritesCount)
				}
				if article.Author.Username != "testuser" {
					t.Errorf("Expected author username %q, got %q", "testuser", article.Author.Username)
				}
				if article.Author.Bio != "Test Bio" {
					t.Errorf("Expected author bio to be %q, got %q", "Test Bio", article.Author.Bio)
				}
				if article.Author.Image != "https://example.com/image.jpg" {
					t.Errorf("Expected author image to be %q, got %q", "https://example.com/image.jpg", article.Author.Image)
				}
				if article.Author.Following {
					t.Errorf("Expected author following to be false, got true")
				}
			},
		},
		{
			name:        "User not found",
			userID:      999,
			title:       "Test Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{"tag1", "tag2"},
			setupMock: func() *MockArticleRepository {
				return &MockArticleRepository{
					createFunc: func(userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						return nil, repository.ErrUserNotFound
					},
				}
			},
			expectedErr: ErrUserNotFound,
			validate:    nil,
		},
		{
			name:        "Duplicate slug",
			userID:      1,
			title:       "Duplicate Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{"tag1", "tag2"},
			setupMock: func() *MockArticleRepository {
				return &MockArticleRepository{
					createFunc: func(userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						return nil, repository.ErrDuplicateSlug
					},
				}
			},
			expectedErr: ErrArticleAlreadyExists,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockArticleRepository := tt.setupMock()

			// Create service with mock repository
			articleService := NewArticleService(mockArticleRepository)

			// Call method
			article, err := articleService.CreateArticle(tt.userID, tt.title, tt.description, tt.body, tt.tagList)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate article if expected
			if err == nil && tt.validate != nil {
				tt.validate(t, article)
			}
		})
	}
}
