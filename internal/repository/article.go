package repository

// Article represents an article in the repository
type Article struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	Body        string
	Author      *User
	CreatedAt   string
	UpdatedAt   string
	TagList     []string
}
