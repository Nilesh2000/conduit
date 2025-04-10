package repository

// Article represents an article in the system
type Article struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	Body        string
	AuthorID    int64
	CreatedAt   string
	UpdatedAt   string
	TagList     []string
}
