package repository

import "time"

// Article represents an article in the repository
type Article struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	Body        string
	AuthorID    int64
	Author      *User
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TagList     []string
}
