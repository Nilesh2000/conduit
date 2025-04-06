package main

import (
	"time"
)

type LoginUser struct {
	Email    string
	Password string
}

type NewUser struct {
	Username string
	Email    string
	Password string
}

type User struct {
	Email    string
	Token    string
	Username string
	Bio      string
	Image    string
}

type UpdateUser struct {
	Email    string
	Token    string
	Username string
	Bio      string
	Image    string
}

type Profile struct {
	Username  string
	Bio       string
	Image     string
	Following bool
}

type Article struct {
	Slug           string
	Title          string
	Description    string
	Body           string
	TagList        []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Favorited      bool
	FavoritesCount int
	Author         Profile
}

type NewArticle struct {
	Title       string
	Description string
	Body        string
	TagList     []string
}

type UpdateArticle struct {
	Title       string
	Description string
	Body        string
}

type Comment struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Body      string
	Author    Profile
}

type NewComment struct {
	Body string
}

type GenericErrorModel struct {
	Errors struct {
		Body []string
	}
}

type UserService interface {
	Register(username, email, password string) (*User, error)
}
