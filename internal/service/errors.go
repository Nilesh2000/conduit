package service

import "errors"

// Service errors
var (
	ErrUsernameTaken  = errors.New("username already taken")
	ErrEmailTaken     = errors.New("email already registered")
	ErrInternalServer = errors.New("internal server error")

	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
