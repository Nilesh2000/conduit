package repository

import "errors"

// Repository errors
var (
	ErrDuplicateUsername = errors.New("username already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrInternal          = errors.New("internal repository error")

	ErrUserNotFound = errors.New("user not found")

	ErrDuplicateSlug = errors.New("article slug already exists")
)
