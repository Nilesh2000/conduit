package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"

	"github.com/gosimple/slug"
)

// MockArticleRepository is a mock implementation of the ArticleRepository interface
type MockArticleRepository struct {
	createFunc            func(ctx context.Context, userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error)
	getBySlugFunc         func(ctx context.Context, slug string) (*repository.Article, error)
	updateFunc            func(ctx context.Context, userID int64, slug string, title, description, body *string) (*repository.Article, error)
	deleteFunc            func(ctx context.Context, articleID int64) error
	favoriteFunc          func(ctx context.Context, userID int64, articleID int64) error
	unfavoriteFunc        func(ctx context.Context, userID int64, articleID int64) error
	getFavoritesCountFunc func(ctx context.Context, articleID int64) (int, error)
	isFavoritedFunc       func(ctx context.Context, userID int64, articleID int64) (bool, error)
	listArticlesFunc      func(ctx context.Context, filters repository.ArticleFilters, currentUserID *int64) (*repository.ArticleListResult, error)
	getArticlesFeedFunc   func(ctx context.Context, userID int64, limit, offset int) (*repository.ArticleListResult, error)
}

// Create is a mock implementation of the Create method
func (m *MockArticleRepository) Create(
	ctx context.Context,
	userID int64,
	articleSlug, title, description, body string,
	tagList []string,
) (*repository.Article, error) {
	return m.createFunc(ctx, userID, articleSlug, title, description, body, tagList)
}

// GetBySlug is a mock implementation of the GetBySlug method
func (m *MockArticleRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*repository.Article, error) {
	return m.getBySlugFunc(ctx, slug)
}

// Update is a mock implementation of the Update method
func (m *MockArticleRepository) Update(
	ctx context.Context,
	userID int64,
	slug string,
	title, description, body *string,
) (*repository.Article, error) {
	return m.updateFunc(ctx, userID, slug, title, description, body)
}

// Delete is a mock implementation of the Delete method
func (m *MockArticleRepository) Delete(
	ctx context.Context,
	articleID int64,
) error {
	return m.deleteFunc(ctx, articleID)
}

// Favorite is a mock implementation of the Favorite method
func (m *MockArticleRepository) Favorite(
	ctx context.Context,
	userID int64,
	articleID int64,
) error {
	return m.favoriteFunc(ctx, userID, articleID)
}

// Unfavorite is a mock implementation of the Unfavorite method
func (m *MockArticleRepository) Unfavorite(
	ctx context.Context,
	userID int64,
	articleID int64,
) error {
	return m.unfavoriteFunc(ctx, userID, articleID)
}

// GetFavoritesCount is a mock implementation of the GetFavoritesCount method
func (m *MockArticleRepository) GetFavoritesCount(
	ctx context.Context,
	articleID int64,
) (int, error) {
	return m.getFavoritesCountFunc(ctx, articleID)
}

// IsFavorited is a mock implementation of the IsFavorited method
func (m *MockArticleRepository) IsFavorited(
	ctx context.Context,
	userID int64,
	articleID int64,
) (bool, error) {
	return m.isFavoritedFunc(ctx, userID, articleID)
}

// ListArticles is a mock implementation of the ListArticles method
func (m *MockArticleRepository) ListArticles(
	ctx context.Context,
	filters repository.ArticleFilters,
	currentUserID *int64,
) (*repository.ArticleListResult, error) {
	return m.listArticlesFunc(ctx, filters, currentUserID)
}

// GetArticlesFeed is a mock implementation of the GetArticlesFeed method
func (m *MockArticleRepository) GetArticlesFeed(
	ctx context.Context,
	userID int64,
	limit, offset int,
) (*repository.ArticleListResult, error) {
	return m.getArticlesFeedFunc(ctx, userID, limit, offset)
}

// Test_articleService_CreateArticle tests the CreateArticle method of the articleService
func Test_articleService_CreateArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      int64
		title       string
		description string
		body        string
		tagList     []string
		setupMock   func() (*MockArticleRepository, *MockProfileRepository)
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
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					createFunc: func(ctx context.Context, userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						expectedSlug := slug.Make("Test Article")
						if articleSlug != expectedSlug {
							t.Errorf("Expected slug %q, got %q", expectedSlug, articleSlug)
						}
						if title != "Test Article" {
							t.Errorf("Expected title %q, got %q", "Test Article", title)
						}
						if description != "Test Description" {
							t.Errorf(
								"Expected description %q, got %q",
								"Test Description",
								description,
							)
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
				return mockArticleRepo, nil
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
					t.Errorf(
						"Expected description %q, got %q",
						"Test Description",
						article.Description,
					)
				}
				if article.Body != "Test Body" {
					t.Errorf("Expected body %q, got %q", "Test Body", article.Body)
				}
				if len(article.TagList) != 2 || article.TagList[0] != "tag1" ||
					article.TagList[1] != "tag2" {
					t.Errorf("Expected tags %v, got %v", []string{"tag1", "tag2"}, article.TagList)
				}
				if article.Favorited {
					t.Errorf("Expected favorited to be false, got true")
				}
				if article.FavoritesCount != 0 {
					t.Errorf("Expected favorites count to be 0, got %d", article.FavoritesCount)
				}
				if article.Author.Username != "testuser" {
					t.Errorf(
						"Expected author username %q, got %q",
						"testuser",
						article.Author.Username,
					)
				}
				if article.Author.Bio != "Test Bio" {
					t.Errorf("Expected author bio to be %q, got %q", "Test Bio", article.Author.Bio)
				}
				if article.Author.Image != "https://example.com/image.jpg" {
					t.Errorf(
						"Expected author image to be %q, got %q",
						"https://example.com/image.jpg",
						article.Author.Image,
					)
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
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				return &MockArticleRepository{
					createFunc: func(ctx context.Context, userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						return nil, repository.ErrUserNotFound
					},
				}, nil
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
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				return &MockArticleRepository{
					createFunc: func(ctx context.Context, userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						return nil, repository.ErrDuplicateSlug
					},
				}, nil
			},
			expectedErr: ErrArticleAlreadyExists,
			validate:    nil,
		},
		{
			name:        "Internal server error",
			userID:      1,
			title:       "Test Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{"tag1", "tag2"},
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				return &MockArticleRepository{
					createFunc: func(ctx context.Context, userID int64, articleSlug, title, description, body string, tagList []string) (*repository.Article, error) {
						return nil, repository.ErrInternal
					},
				}, nil
			},
			expectedErr: ErrInternalServer,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockArticleRepository, mockProfileRepository := tt.setupMock()

			// Create service with mock repository
			articleService := NewArticleService(mockArticleRepository, mockProfileRepository)

			// Create context
			ctx := context.Background()

			// Call method
			article, err := articleService.CreateArticle(
				ctx,
				tt.userID,
				tt.title,
				tt.description,
				tt.body,
				tt.tagList,
			)

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

// Test_articleService_GetArticle tests the GetArticle method of the articleService
func Test_articleService_GetArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		slug        string
		setupMock   func() (*MockArticleRepository, *MockProfileRepository)
		expectedErr error
		validate    func(*testing.T, *Article)
	}{
		{
			name: "Article found",
			slug: "test-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						if slug != "test-article" {
							t.Errorf("Expected slug %q, got %q", "test-article", slug)
						}

						return &repository.Article{
							ID:          1,
							Slug:        "test-article",
							Title:       "Test Article",
							Description: "Test Description",
							Body:        "Test Body",
							AuthorID:    1,
							Author: &repository.User{
								ID:       1,
								Username: "testuser",
								Bio:      "Test Bio",
								Image:    "https://example.com/image.jpg",
							},
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
							TagList:   []string{"tag1", "tag2"},
						}, nil
					},
					getFavoritesCountFunc: func(ctx context.Context, articleID int64) (int, error) {
						return 0, nil
					},
					isFavoritedFunc: func(ctx context.Context, userID int64, articleID int64) (bool, error) {
						return false, nil
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
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
					t.Errorf(
						"Expected description %q, got %q",
						"Test Description",
						article.Description,
					)
				}
				if article.Body != "Test Body" {
					t.Errorf("Expected body %q, got %q", "Test Body", article.Body)
				}
				if len(article.TagList) != 2 || article.TagList[0] != "tag1" ||
					article.TagList[1] != "tag2" {
					t.Errorf("Expected tags %v, got %v", []string{"tag1", "tag2"}, article.TagList)
				}
				if article.Author.Username != "testuser" {
					t.Errorf(
						"Expected author username %q, got %q",
						"testuser",
						article.Author.Username,
					)
				}
				if article.Author.Bio != "Test Bio" {
					t.Errorf("Expected author bio %q, got %q", "Test Bio", article.Author.Bio)
				}
				if article.Author.Image != "https://example.com/image.jpg" {
					t.Errorf(
						"Expected author image %q, got %q",
						"https://example.com/image.jpg",
						article.Author.Image,
					)
				}
				if article.Favorited {
					t.Errorf("Expected favorited to be false, got true")
				}
				if article.FavoritesCount != 0 {
					t.Errorf("Expected favorites count to be 0, got %d", article.FavoritesCount)
				}
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						return nil, repository.ErrArticleNotFound
					},
					getFavoritesCountFunc: func(ctx context.Context, articleID int64) (int, error) {
						return 0, nil
					},
					isFavoritedFunc: func(ctx context.Context, userID int64, articleID int64) (bool, error) {
						return false, nil
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: ErrArticleNotFound,
			validate:    nil,
		},
		{
			name: "Repository error",
			slug: "test-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						return nil, repository.ErrInternal
					},
					getFavoritesCountFunc: func(ctx context.Context, articleID int64) (int, error) {
						return 0, nil
					},
					isFavoritedFunc: func(ctx context.Context, userID int64, articleID int64) (bool, error) {
						return false, nil
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: ErrInternalServer,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockArticleRepository, mockProfileRepository := tt.setupMock()

			// Create service with mock repository
			articleService := NewArticleService(mockArticleRepository, mockProfileRepository)

			// Create context
			ctx := context.Background()

			// Call method
			article, err := articleService.GetArticle(ctx, tt.slug, nil)

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

// Test_articleService_FavoriteArticle tests the FavoriteArticle method of the articleService
func Test_articleService_FavoriteArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      int64
		slug        string
		setupMock   func() (*MockArticleRepository, *MockProfileRepository)
		expectedErr error
		validate    func(*testing.T, *Article)
	}{
		{
			name:   "Successfully favorite an article",
			userID: 1,
			slug:   "test-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						now := time.Now()
						return &repository.Article{
							ID:          1,
							Slug:        "test-article",
							Title:       "Test Article",
							Description: "Test Description",
							Body:        "Test Body",
							AuthorID:    2,
							Author: &repository.User{
								ID:       2,
								Username: "authoruser",
								Bio:      "Author Bio",
								Image:    "https://example.com/image.jpg",
							},
							CreatedAt: now,
							UpdatedAt: now,
							TagList:   []string{"tag1", "tag2"},
						}, nil
					},
					favoriteFunc: func(ctx context.Context, userID int64, articleID int64) error {
						if userID != 1 || articleID != 1 {
							t.Errorf(
								"Expected FavoriteArticle(1, 1), got FavoriteArticle(%d, %d)",
								userID,
								articleID,
							)
						}
						return nil
					},
					getFavoritesCountFunc: func(ctx context.Context, articleID int64) (int, error) {
						return 1, nil
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: nil,
			validate: func(t *testing.T, article *Article) {
				if article == nil {
					t.Errorf("Expected article to be not nil, got nil")
					return
				}
				if !article.Favorited {
					t.Errorf("Expected favorited to be true, got false")
				}
				if article.FavoritesCount != 1 {
					t.Errorf("Expected favorites count to be 1, got %d", article.FavoritesCount)
				}
			},
		},
		{
			name:   "Article not found",
			userID: 1,
			slug:   "non-existent-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						return nil, repository.ErrArticleNotFound
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: ErrArticleNotFound,
			validate:    nil,
		},
		{
			name:   "Failed to favorite an article",
			userID: 1,
			slug:   "test-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						return &repository.Article{
							ID:          1,
							Slug:        "test-article",
							Title:       "Test Article",
							Description: "Test Description",
							Body:        "Test Body",
							AuthorID:    2,
							Author: &repository.User{
								ID:       2,
								Username: "authoruser",
								Bio:      "Author Bio",
								Image:    "https://example.com/image.jpg",
							},
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
							TagList:   []string{"tag1", "tag2"},
						}, nil
					},
					favoriteFunc: func(ctx context.Context, userID int64, articleID int64) error {
						return repository.ErrInternal
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: ErrInternalServer,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockArticleRepository, mockProfileRepository := tt.setupMock()

			// Create service with mock repository
			articleService := NewArticleService(mockArticleRepository, mockProfileRepository)

			// Create context
			ctx := context.Background()

			// Call method
			article, err := articleService.FavoriteArticle(ctx, tt.userID, tt.slug)

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

// Test_articleService_UnfavoriteArticle tests the UnfavoriteArticle method of the articleService
func Test_articleService_UnfavoriteArticle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		userID      int64
		slug        string
		setupMock   func() (*MockArticleRepository, *MockProfileRepository)
		expectedErr error
		validate    func(*testing.T, *Article)
	}{
		{
			name:   "Successfully unfavorite an article",
			userID: 1,
			slug:   "test-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						now := time.Now()
						return &repository.Article{
							ID:          1,
							Slug:        "test-article",
							Title:       "Test Article",
							Description: "Test Description",
							Body:        "Test Body",
							AuthorID:    2,
							Author: &repository.User{
								ID:       2,
								Username: "authoruser",
								Bio:      "Author Bio",
								Image:    "https://example.com/image.jpg",
							},
							CreatedAt: now,
							UpdatedAt: now,
							TagList:   []string{"tag1", "tag2"},
						}, nil
					},
					unfavoriteFunc: func(ctx context.Context, userID int64, articleID int64) error {
						if userID != 1 || articleID != 1 {
							t.Errorf(
								"Expected UnfavoriteArticle(1, 1), got UnfavoriteArticle(%d, %d)",
								userID,
								articleID,
							)
						}
						return nil
					},
					getFavoritesCountFunc: func(ctx context.Context, articleID int64) (int, error) {
						return 0, nil
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: nil,
			validate: func(t *testing.T, article *Article) {
				if article == nil {
					t.Errorf("Expected article to be not nil, got nil")
					return
				}
				if article.Favorited {
					t.Errorf("Expected favorited to be false, got true")
				}
				if article.FavoritesCount != 0 {
					t.Errorf("Expected favorites count to be 0, got %d", article.FavoritesCount)
				}
			},
		},
		{
			name:   "Article not found",
			userID: 1,
			slug:   "non-existent-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						return nil, repository.ErrArticleNotFound
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: ErrArticleNotFound,
			validate:    nil,
		},
		{
			name:   "Failed to unfavorite an article",
			userID: 1,
			slug:   "test-article",
			setupMock: func() (*MockArticleRepository, *MockProfileRepository) {
				mockArticleRepo := &MockArticleRepository{
					getBySlugFunc: func(ctx context.Context, slug string) (*repository.Article, error) {
						now := time.Now()
						return &repository.Article{
							ID:          1,
							Slug:        "test-article",
							Title:       "Test Article",
							Description: "Test Description",
							Body:        "Test Body",
							AuthorID:    2,
							Author: &repository.User{
								ID:       2,
								Username: "authoruser",
								Bio:      "Author Bio",
								Image:    "https://example.com/image.jpg",
							},
							CreatedAt: now,
							UpdatedAt: now,
							TagList:   []string{"tag1", "tag2"},
						}, nil
					},
					unfavoriteFunc: func(ctx context.Context, userID int64, articleID int64) error {
						return repository.ErrInternal
					},
				}
				mockProfileRepo := &MockProfileRepository{
					isFollowingFunc: func(ctx context.Context, followerID int64, followingID int64) (bool, error) {
						return false, nil
					},
				}
				return mockArticleRepo, mockProfileRepo
			},
			expectedErr: ErrInternalServer,
			validate:    nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock repository
			mockArticleRepository, mockProfileRepository := tt.setupMock()

			// Create service with mock repository
			articleService := NewArticleService(mockArticleRepository, mockProfileRepository)

			// Create context
			ctx := context.Background()

			// Call method
			article, err := articleService.UnfavoriteArticle(ctx, tt.userID, tt.slug)

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
