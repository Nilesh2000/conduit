package repository

import "errors"

var (
	ErrDuplicateUsername = errors.New("username already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrInternal          = errors.New("internal repository error")
)
