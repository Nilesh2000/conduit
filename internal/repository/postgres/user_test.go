package postgres

import (
	"conduit/internal/repository"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

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
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO users \(username, email, password, bio, image, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\) RETURNING id`).
					WithArgs("testuser", "test@example.com", "hashedPassword", "", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

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
			},
		},
		{
			name:     "Duplicate username",
			username: "existinguser",
			email:    "new@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO users \(username, email, password, bio, image, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\) RETURNING id`).
					WithArgs("existinguser", "new@example.com", "hashedPassword", "", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
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
			username: "newuser",
			email:    "existing@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO users \(username, email, password, bio, image, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\) RETURNING id`).
					WithArgs("newuser", "existing@example.com", "hashedPassword", "", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
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
			name:     "Database Error",
			username: "testuser",
			email:    "test@example.com",
			password: "hashedPassword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO users \(username, email, password, bio, image, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\) RETURNING id`).
					WithArgs("testuser", "test@example.com", "hashedPassword", "", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))

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
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO users \(username, email, password, bio, image, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\) RETURNING id`).
					WithArgs("testuser", "test@example.com", "hashedPassword", "", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectedErr:  repository.ErrInternal,
			validateUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			user, err := repo.Create(tt.username, tt.email, tt.password)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			if err == nil && tt.validateUser != nil {
				tt.validateUser(t, user)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled mock expectations: %v", err)
			}
		})
	}
}
