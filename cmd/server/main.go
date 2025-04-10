package main

import (
	"conduit/internal/config"
	"conduit/internal/handler"
	"conduit/internal/middleware"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	userRepository := postgres.New(db)
	userService := service.NewUserService(userRepository, cfg.JWT.SecretKey, cfg.JWT.Expiry)
	userHandler := handler.NewUserHandler(userService)

	router := http.NewServeMux()
	router.HandleFunc("POST /api/users", userHandler.Register())
	router.HandleFunc("POST /api/users/login", userHandler.Login())

	router.HandleFunc("GET /api/user", middleware.RequireAuth([]byte(cfg.JWT.SecretKey))(userHandler.GetCurrentUser()))
	router.HandleFunc("PUT /api/user", middleware.RequireAuth([]byte(cfg.JWT.SecretKey))(userHandler.UpdateCurrentUser()))

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	log.Printf("Starting server on :%s", cfg.Server.Port)
	log.Fatal(server.ListenAndServe())
}
