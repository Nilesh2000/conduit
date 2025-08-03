package repository

import "time"

// Article represents an article in the repository
type Article struct {
	ID             int64
	Slug           string
	Title          string
	Description    string
	Body           string
	AuthorID       int64
	Author         *User
	CreatedAt      time.Time
	UpdatedAt      time.Time
	TagList        []string
	Favorited      bool
	FavoritesCount int
}

// ArticleFilters represents filters for listing articles
type ArticleFilters struct {
	Tag       *string
	Author    *string
	Favorited *string
	Limit     int
	Offset    int
}

// ArticleListResult represents the result of listing articles
type ArticleListResult struct {
	Articles []*Article
	Count    int
}
