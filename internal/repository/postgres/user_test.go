package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"conduit/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

// setupTestDB creates a new mock database for testing
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	return db, mock
}

// TestCreate tests the Create method of the UserRepository
func Test_userRepository_Create(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name         string
		username     string
		email        string
		password     string
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedErr  error
		validateUser func(*testing.T, *repository.User)
	}{
		{
			name:     "Successful creation",
			username: "testuser",
			email:    "test@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect insert query with returning id
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs("testuser", "test@example.com", "hashedPassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).AddRow(1, "testuser", "test@example.com", "hashedPassword", nil, nil, time.Now(), time.Now()))

				// Expect commit
				mock.ExpectCommit()
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.ID != 1 {
					t.Errorf("Expected ID 1, got %d", user.ID)
				}
				if user.Username != "testuser" {
					t.Errorf("Expected username testuser, got %q", user.Username)
				}
				if user.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %q", user.Email)
				}
				if user.Password != "hashedPassword" {
					t.Errorf("Expected password hashedPassword, got %q", user.Password)
				}
				if user.Bio != "" {
					t.Errorf("Expected empty bio, got %q", user.Bio)
				}
				if user.Image != "" {
					t.Errorf("Expected empty image, got %q", user.Image)
				}
				if user.CreatedAt.IsZero() {
					t.Errorf("Expected non-zero created_at, got %v", user.CreatedAt)
				}
				if user.UpdatedAt.IsZero() {
					t.Errorf("Expected non-zero updated_at, got %v", user.UpdatedAt)
				}
			},
		},
		{
			name:     "Duplicate username",
			username: "existinguser",
			email:    "new@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect insert query to fail with duplicate key error on username
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, username, email, password`).
					WithArgs("existinguser", "new@example.com", "hashedPassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(&pq.Error{
						Code:       "23505",
						Message:    "duplicate key value violates unique constraint",
						Constraint: "users_username_key",
					})

				// Expect rollback
				mock.ExpectRollback()
			},
			expectedErr:  repository.ErrDuplicateUsername,
			validateUser: nil,
		},
		{
			name:     "Duplicate email",
			username: "newuser",
			email:    "existing@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect insert query to fail with duplicate key error on email
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, username, email, password`).
					WithArgs("newuser", "existing@example.com", "hashedPassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(&pq.Error{
						Code:       "23505",
						Message:    "duplicate key value violates unique constraint",
						Constraint: "users_email_key",
					})

				// Expect rollback
				mock.ExpectRollback()
			},
			expectedErr:  repository.ErrDuplicateEmail,
			validateUser: nil,
		},
		{
			name:     "Database Error",
			username: "testuser",
			email:    "test@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect insert query to fail with generic database error
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, username, email, password`).
					WithArgs("testuser", "test@example.com", "hashedPassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))

				// Expect rollback
				mock.ExpectRollback()
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name:     "Transaction Begin Error",
			username: "testuser",
			email:    "test@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name:     "Transaction Commit Error",
			username: "testuser",
			email:    "test@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect insert query with returning id
				mock.ExpectQuery(`INSERT INTO users \(username, email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs("testuser", "test@example.com", "hashedPassword", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).AddRow(1, "testuser", "test@example.com", "hashedPassword", nil, nil, time.Now(), time.Now()))

				// Expect commit transaction to fail
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
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
			repo := NewUserRepository(db)

			// Create context
			ctx := context.Background()

			// Call Create method
			user, err := repo.Create(ctx, tt.username, tt.email, tt.password)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate user if no error
			if err == nil && tt.validateUser != nil {
				tt.validateUser(t, user)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestFindByEmail tests the FindByEmail method of the UserRepository
func Test_userRepository_FindByEmail(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name         string
		email        string
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedErr  error
		validateUser func(*testing.T, *repository.User)
	}{
		{
			name:  "User found",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedPassword", "Test bio", "test.jpg", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.ID != 1 {
					t.Errorf("Expected ID 1, got %d", user.ID)
				}
				if user.Username != "testuser" {
					t.Errorf("Expected username testuser, got %q", user.Username)
				}
				if user.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %q", user.Email)
				}
				if user.Bio != "Test bio" {
					t.Errorf("Expected bio 'Test bio', got %q", user.Bio)
				}
				if user.Image != "test.jpg" {
					t.Errorf("Expected image 'test.jpg', got %q", user.Image)
				}
			},
		},
		{
			name:  "User not found",
			email: "nonexistent@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr:  repository.ErrUserNotFound,
			validateUser: nil,
		},
		{
			name:  "Database Error",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnError(errors.New("database error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name:  "Null bio and image",
			email: "nullfields@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "testuser", "nullfields@example.com", "hashedPassword", nil, nil, time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nullfields@example.com").
					WillReturnRows(rows)
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.Bio != "" {
					t.Errorf("Expected empty bio, got %q", user.Bio)
				}
				if user.Image != "" {
					t.Errorf("Expected empty image, got %q", user.Image)
				}
			},
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
			repo := NewUserRepository(db)

			// Create context
			ctx := context.Background()

			// Call FindByEmail method
			user, err := repo.FindByEmail(ctx, tt.email)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate user if no error
			if err == nil && tt.validateUser != nil {
				tt.validateUser(t, user)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestFindByID tests the FindByID method of the UserRepository
func Test_userRepository_FindByID(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name         string
		id           int64
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedErr  error
		validateUser func(*testing.T, *repository.User)
	}{
		{
			name: "User found",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedPassword", "Test bio", "test.jpg", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.ID != 1 {
					t.Errorf("Expected ID 1, got %d", user.ID)
				}
				if user.Username != "testuser" {
					t.Errorf("Expected username testuser, got %q", user.Username)
				}
				if user.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %q", user.Email)
				}
				if user.Bio != "Test bio" {
					t.Errorf("Expected bio 'Test bio', got %q", user.Bio)
				}
				if user.Image != "test.jpg" {
					t.Errorf("Expected image 'test.jpg', got %q", user.Image)
				}
				if user.CreatedAt.IsZero() {
					t.Errorf("Expected non-zero created_at, got %v", user.CreatedAt)
				}
				if user.UpdatedAt.IsZero() {
					t.Errorf("Expected non-zero updated_at, got %v", user.UpdatedAt)
				}
			},
		},
		{
			name: "User not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr:  repository.ErrUserNotFound,
			validateUser: nil,
		},
		{
			name: "Database Error",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name: "Null bio and image",
			id:   2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(2, "testuser", "test@example.com", "hashedPassword", nil, nil, time.Now(), time.Now())
				mock.ExpectQuery(`SELECT id, username, email, password, bio, image, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.Bio != "" {
					t.Errorf("Expected empty bio, got %q", user.Bio)
				}
				if user.Image != "" {
					t.Errorf("Expected empty image, got %q", user.Image)
				}
			},
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
			repo := NewUserRepository(db)

			// Create context
			ctx := context.Background()

			// Call FindByID method
			user, err := repo.FindByID(ctx, tt.id)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate user if no error
			if err == nil && tt.validateUser != nil {
				tt.validateUser(t, user)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

// TestUpdate tests the Update method of the UserRepository
func Test_userRepository_Update(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string {
		return &s
	}

	// Define test cases
	tests := []struct {
		name         string
		userID       int64
		username     *string
		email        *string
		password     *string
		bio          *string
		image        *string
		mockSetup    func(mock sqlmock.Sqlmock)
		expectedErr  error
		validateUser func(*testing.T, *repository.User)
	}{
		{
			name:     "Successful update all fields",
			userID:   1,
			username: strPtr("updateduser"),
			email:    strPtr("updated@example.com"),
			password: strPtr("updatedpassword"),
			bio:      strPtr("Updated bio"),
			image:    strPtr("updatedimage.jpg"),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "updateduser", "updated@example.com", "updatedpassword", "Updated bio", "updatedimage.jpg", time.Now(), time.Now())

				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(strPtr("updateduser"), strPtr("updated@example.com"), strPtr("updatedpassword"), strPtr("Updated bio"), strPtr("updatedimage.jpg"), sqlmock.AnyArg(), 1).
					WillReturnRows(rows)

				mock.ExpectCommit()
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.Username != "updateduser" {
					t.Errorf("Expected username updateduser, got %q", user.Username)
				}
				if user.Email != "updated@example.com" {
					t.Errorf("Expected email updated@example.com, got %q", user.Email)
				}
				if user.Password != "updatedpassword" {
					t.Errorf("Expected password updatedpassword, got %q", user.Password)
				}
				if user.Bio != "Updated bio" {
					t.Errorf("Expected bio Updated bio, got %q", user.Bio)
				}
				if user.Image != "updatedimage.jpg" {
					t.Errorf("Expected image updatedimage.jpg, got %q", user.Image)
				}
				if user.CreatedAt.IsZero() {
					t.Errorf("Expected non-zero created_at, got %v", user.CreatedAt)
				}
				if user.UpdatedAt.IsZero() {
					t.Errorf("Expected non-zero updated_at, got %v", user.UpdatedAt)
				}
			},
		},
		{
			name:     "Successful update partial only",
			userID:   1,
			username: nil,
			email:    nil,
			password: nil,
			bio:      strPtr("New bio only"),
			image:    strPtr("new-image.jpg"),
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedPassword", "New bio only", "new-image.jpg", time.Now(), time.Now())

				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(nil, nil, nil, strPtr("New bio only"), strPtr("new-image.jpg"), sqlmock.AnyArg(), 1).
					WillReturnRows(rows)

				mock.ExpectCommit()
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.Username != "testuser" {
					t.Errorf("Expected username testuser, got %q", user.Username)
				}
				if user.Email != "test@example.com" {
					t.Errorf("Expected email test@example.com, got %q", user.Email)
				}
				if user.Password != "hashedPassword" {
					t.Errorf("Expected password hashedPassword, got %q", user.Password)
				}
				if user.Bio != "New bio only" {
					t.Errorf("Expected bio New bio only, got %q", user.Bio)
				}
				if user.Image != "new-image.jpg" {
					t.Errorf("Expected image new-image.jpg, got %q", user.Image)
				}
				if user.CreatedAt.IsZero() {
					t.Errorf("Expected non-zero created_at, got %v", user.CreatedAt)
				}
				if user.UpdatedAt.IsZero() {
					t.Errorf("Expected non-zero updated_at, got %v", user.UpdatedAt)
				}
			},
		},
		{
			name:     "Duplicate username",
			userID:   1,
			username: strPtr("existinguser"),
			email:    nil,
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(strPtr("existinguser"), nil, nil, nil, nil, sqlmock.AnyArg(), 1).
					WillReturnError(&pq.Error{
						Code:       "23505",
						Message:    "duplicate key value violates unique constraint",
						Constraint: "users_username_key",
					})
				mock.ExpectRollback()
			},
			expectedErr:  repository.ErrDuplicateUsername,
			validateUser: nil,
		},
		{
			name:     "Duplicate email",
			userID:   1,
			username: nil,
			email:    strPtr("existingemail@example.com"),
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(nil, strPtr("existingemail@example.com"), nil, nil, nil, sqlmock.AnyArg(), 1).
					WillReturnError(&pq.Error{
						Code:       "23505",
						Message:    "duplicate key value violates unique constraint",
						Constraint: "users_email_key",
					})
				mock.ExpectRollback()
			},
			expectedErr:  repository.ErrDuplicateEmail,
			validateUser: nil,
		},
		{
			name:     "User not found",
			userID:   1,
			username: strPtr("newname"),
			email:    nil,
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(strPtr("newname"), nil, nil, nil, nil, sqlmock.AnyArg(), 1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedErr:  repository.ErrUserNotFound,
			validateUser: nil,
		},
		{
			name:     "Database error",
			userID:   1,
			username: strPtr("newname"),
			email:    nil,
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(strPtr("newname"), nil, nil, nil, nil, sqlmock.AnyArg(), 1).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name:     "Transaction begin error",
			userID:   1,
			username: strPtr("newname"),
			email:    nil,
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("transaction error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name:     "Transaction commit error",
			userID:   1,
			username: strPtr("newname"),
			email:    nil,
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedPassword", "New bio only", "new-image.jpg", time.Now(), time.Now())
				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(strPtr("newname"), nil, nil, nil, nil, sqlmock.AnyArg(), 1).
					WillReturnRows(rows)
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
		{
			name:     "Update with null bio and image",
			userID:   1,
			username: strPtr("newname"),
			email:    nil,
			password: nil,
			bio:      nil,
			image:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "bio", "image", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedPassword", nil, nil, time.Now(), time.Now())
				mock.ExpectQuery(`UPDATE users SET username = COALESCE\(\$1, username\), email = COALESCE\(\$2, email\), password = COALESCE\(\$3, password\), bio = COALESCE\(\$4, bio\), image = COALESCE\(\$5, image\), updated_at = \$6 WHERE id = \$7 RETURNING id, username, email, password, bio, image, created_at, updated_at`).
					WithArgs(strPtr("newname"), nil, nil, nil, nil, sqlmock.AnyArg(), 1).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
			expectedErr: nil,
			validateUser: func(t *testing.T, user *repository.User) {
				if user.Bio != "" {
					t.Errorf("Expected empty bio, got %q", user.Bio)
				}
				if user.Image != "" {
					t.Errorf("Expected empty image, got %q", user.Image)
				}
			},
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
			repo := NewUserRepository(db)

			// Create context
			ctx := context.Background()

			// Call Update method
			user, err := repo.Update(
				ctx,
				tt.userID,
				tt.username,
				tt.email,
				tt.password,
				tt.bio,
				tt.image,
			)

			// Validate error
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			// Validate user if no error
			if err == nil && tt.validateUser != nil {
				tt.validateUser(t, user)
			}

			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}
