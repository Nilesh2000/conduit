package service

import "errors"

var (
	ErrUsernameTaken  = errors.New("username already taken")
	ErrEmailTaken     = errors.New("email already registered")
	ErrInternalServer = errors.New("internal server error")
)
