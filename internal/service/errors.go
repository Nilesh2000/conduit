package service

import "errors"

// Service errors
var (
	ErrUsernameTaken  = errors.New("username already taken")
	ErrEmailTaken     = errors.New("email already registered")
	ErrInternalServer = errors.New("internal server error")

	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrArticleAlreadyExists = errors.New("article with this title already exists")

	ErrArticleNotFound = errors.New("article not found")

	ErrCannotFollowSelf = errors.New("cannot follow yourself")

	ErrArticleAuthorCannotFavorite = errors.New("article author cannot favorite their own article")

	ErrArticleNotAuthorized = errors.New("article not authorized")
)
