package main

import (
	"conduit/internal/handler"
	"conduit/internal/repository/postgres"
	"conduit/internal/service"
	"database/sql"
	"log"
	"net/http"
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

func main() {
	connStr := "postgres://postgres:admin@localhost:5432/conduit?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepository := postgres.NewUserRepository(db)
	userService := service.NewUserService(userRepository, "secret", time.Hour*24)
	userHandler := handler.NewUserHandler(userService)

	router := http.NewServeMux()
	router.HandleFunc("POST /api/users", userHandler.Register())

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
}
