package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nilesh2000/conduit/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func Test_articleRepository_Create(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name            string
		userID          int64
		slug            string
		title           string
		description     string
		body            string
		tagList         []string
		mockSetup       func(mock sqlmock.Sqlmock)
		expectedErr     error
		validateArticle func(*testing.T, *repository.Article)
	}{
		{
			name:        "Successful creation",
			userID:      1,
			slug:        "test-article",
			title:       "Test Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{"tag1", "tag2"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock beginning transaction
				mock.ExpectBegin()

				mock.ExpectQuery(`WITH inserted_article AS`).
					WithArgs("test-article", "Test Article", "Test Description", "Test Body", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "title", "description", "body", "author_id", "created_at", "updated_at", "id", "username", "bio", "image"}).
						AddRow(1, "test-article", "Test Article", "Test Description", "Test Body", int64(1), time.Now(), time.Now(), 1, "testuser", "Test Bio", "https://example.com/image.jpg"),
					)

				// Expect insert tag queries
				mock.ExpectQuery(`INSERT INTO tags \(name\) VALUES \(\$1\) ON CONFLICT \(name\) DO UPDATE SET name = EXCLUDED.name RETURNING id`).
					WithArgs("tag1").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectExec(`INSERT INTO article_tags \(article_id, tag_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectQuery(`INSERT INTO tags \(name\) VALUES \(\$1\) ON CONFLICT \(name\) DO UPDATE SET name = EXCLUDED.name RETURNING id`).
					WithArgs("tag2").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

				mock.ExpectExec(`INSERT INTO article_tags \(article_id, tag_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect commit
				mock.ExpectCommit()
			},
			expectedErr: nil,
			validateArticle: func(t *testing.T, article *repository.Article) {
				if article.ID != 1 {
					t.Errorf("Expected ID 1, got %d", article.ID)
				}
				if article.Slug != "test-article" {
					t.Errorf("Expected slug test-article, got %q", article.Slug)
				}
				if article.Title != "Test Article" {
					t.Errorf("Expected title Test Article, got %q", article.Title)
				}
				if article.Description != "Test Description" {
					t.Errorf("Expected description Test Description, got %q", article.Description)
				}
				if article.Body != "Test Body" {
					t.Errorf("Expected body Test Body, got %q", article.Body)
				}
				if article.AuthorID != 1 {
					t.Errorf("Expected author ID 1, got %d", article.AuthorID)
				}
				if article.Author == nil {
					t.Error("Expected author to be populated")
				} else {
					if article.Author.Username != "testuser" {
						t.Errorf("Expected author username testuser, got %q", article.Author.Username)
					}
					if article.Author.Bio != "Test Bio" {
						t.Errorf("Expected author bio Test Bio, got %q", article.Author.Bio)
					}
					if article.Author.Image != "https://example.com/image.jpg" {
						t.Errorf("Expected author image https://example.com/image.jpg, got %q", article.Author.Image)
					}
				}
				if len(article.TagList) != 2 {
					t.Errorf("Expected 2 tags, got %d", len(article.TagList))
				}
				if article.TagList[0] != "tag1" || article.TagList[1] != "tag2" {
					t.Errorf("Expected tags [tag1, tag2], got %v", article.TagList)
				}
				if article.CreatedAt.IsZero() {
					t.Error("Expected created_at to be populated")
				}
				if article.UpdatedAt.IsZero() {
					t.Error("Expected updated_at to be populated")
				}
			},
		},
		{
			name:        "Non-existent user",
			userID:      999,
			slug:        "test-article",
			title:       "Test Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(`WITH inserted_article AS`).
					WithArgs("test-article", "Test Article", "Test Description", "Test Body", int64(999), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(&pq.Error{
						Code:       "23503",
						Message:    "insert or update on table \"articles\" violates foreign key constraint",
						Constraint: "articles_author_id_fkey",
					})

				mock.ExpectRollback()
			},
			expectedErr:     repository.ErrUserNotFound,
			validateArticle: nil,
		},
		{
			name:        "Duplicate slug",
			userID:      1,
			slug:        "existing-article",
			title:       "Test Article",
			description: "Test Description",
			body:        "Test Body",
			tagList:     []string{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(`WITH inserted_article AS`).
					WithArgs("existing-article", "Test Article", "Test Description", "Test Body", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(&pq.Error{
						Code:       "23505",
						Message:    "duplicate key value violates unique constraint",
						Constraint: "articles_slug_key",
					})

				mock.ExpectRollback()
			},
			expectedErr:     repository.ErrDuplicateSlug,
			validateArticle: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock database for this test case
			db, mock := setupTestDB(t)
			defer db.Close()

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			// Create repository with mock database
			repo := NewArticleRepository(db)

			// Create context
			ctx := context.Background()

			// Call Create method
			article, err := repo.Create(
				ctx,
				tt.userID,
				tt.slug,
				tt.title,
				tt.description,
				tt.body,
				tt.tagList,
			)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate article if no error
			if err == nil && tt.validateArticle != nil {
				tt.validateArticle(t, article)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_articleRepository_GetBySlug(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name            string
		slug            string
		mockSetup       func(mock sqlmock.Sqlmock)
		expectedErr     error
		validateArticle func(*testing.T, *repository.Article)
	}{
		{
			name: "Article found",
			slug: "test-article",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "slug", "title", "description", "body", "author_id", "created_at", "updated_at",
					"author_id", "username", "bio", "image",
				}).AddRow(
					1, "test-article", "Test Article", "Test Description", "Test Body", 1,
					time.Now(),
					time.Now(), 1, "testuser", "Test Bio", "https://example.com/image.jpg",
				)

				mock.ExpectQuery(`SELECT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at, u.id, u.username, u.bio, u.image FROM articles a JOIN users u ON a.author_id = u.id WHERE a.slug = \$1`).
					WithArgs("test-article").
					WillReturnRows(rows)

				tagRows := sqlmock.NewRows([]string{"name"}).
					AddRow("tag1").
					AddRow("tag2")

				mock.ExpectQuery(`SELECT t.name FROM tags t JOIN article_tags at ON t.id = at.tag_id WHERE at.article_id = \$1`).
					WithArgs(1).
					WillReturnRows(tagRows)
			},
			expectedErr: nil,
			validateArticle: func(t *testing.T, article *repository.Article) {
				if article.ID != 1 {
					t.Errorf("Expected ID 1, got %d", article.ID)
				}
				if article.Slug != "test-article" {
					t.Errorf("Expected slug test-article, got %q", article.Slug)
				}
				if article.Title != "Test Article" {
					t.Errorf("Expected title Test Article, got %q", article.Title)
				}
				if article.Description != "Test Description" {
					t.Errorf("Expected description Test Description, got %q", article.Description)
				}
				if article.Body != "Test Body" {
					t.Errorf("Expected body Test Body, got %q", article.Body)
				}
				if article.AuthorID != 1 {
					t.Errorf("Expected author ID 1, got %d", article.AuthorID)
				}
				if article.Author == nil {
					t.Error("Expected author to be populated")
				} else {
					if article.Author.ID != 1 {
						t.Errorf("Expected author ID 1, got %d", article.Author.ID)
					}
					if article.Author.Username != "testuser" {
						t.Errorf("Expected author username testuser, got %q", article.Author.Username)
					}
					if article.Author.Bio != "Test Bio" {
						t.Errorf("Expected author bio Test Bio, got %q", article.Author.Bio)
					}
					if article.Author.Image != "https://example.com/image.jpg" {
						t.Errorf("Expected author image https://example.com/image.jpg, got %q", article.Author.Image)
					}
				}
				if len(article.TagList) != 2 {
					t.Errorf("Expected 2 tags, got %d", len(article.TagList))
				}
				if article.TagList[0] != "tag1" || article.TagList[1] != "tag2" {
					t.Errorf("Expected tags [tag1, tag2], got %v", article.TagList)
				}
				if article.CreatedAt.IsZero() {
					t.Error("Expected created_at to be populated")
				}
				if article.UpdatedAt.IsZero() {
					t.Error("Expected updated_at to be populated")
				}
			},
		},
		{
			name: "Article not found",
			slug: "non-existent-article",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at, u.id, u.username, u.bio, u.image FROM articles a JOIN users u ON a.author_id = u.id WHERE a.slug = \$1`).
					WithArgs("non-existent-article").
					WillReturnRows(sqlmock.NewRows([]string{}))
			},
			expectedErr:     repository.ErrArticleNotFound,
			validateArticle: nil,
		},
		{
			name: "Database error",
			slug: "test-article",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT a.id, a.slug, a.title, a.description, a.body, a.author_id, a.created_at, a.updated_at, u.id, u.username, u.bio, u.image FROM articles a JOIN users u ON a.author_id = u.id WHERE a.slug = \$1`).
					WithArgs("test-article").
					WillReturnError(errors.New("database error"))
			},
			expectedErr:     repository.ErrInternal,
			validateArticle: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock database for this test case
			db, mock := setupTestDB(t)
			defer db.Close()

			// Setup mock expectations
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			// Create repository with mock database
			repo := NewArticleRepository(db)

			// Create context
			ctx := context.Background()

			// Call GetBySlug method
			article, err := repo.GetBySlug(ctx, tt.slug)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate article if no error
			if err == nil && tt.validateArticle != nil {
				tt.validateArticle(t, article)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}
