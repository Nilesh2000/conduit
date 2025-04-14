package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"conduit/internal/config"
	"conduit/internal/handler"
	"conduit/internal/middleware"
	"conduit/internal/repository/postgres"
	"conduit/internal/service"
)

type Profile struct {
	Username  string
	Bio       string
	Image     string
	Following bool
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

	// Ping database to check connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize repositories
	userRepository := postgres.NewUserRepository(db)
	articleRepository := postgres.NewArticleRepository(db)
	profileRepository := postgres.NewProfileRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepository, cfg.JWT.SecretKey, cfg.JWT.Expiry)
	articleService := service.NewArticleService(articleRepository)
	profileService := service.NewProfileService(profileRepository)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	articleHandler := handler.NewArticleHandler(articleService)
	profileHandler := handler.NewProfileHandler(profileService)

	// Setup router
	router := http.NewServeMux()

	// Public routes
	router.HandleFunc("POST /api/users", userHandler.Register())
	router.HandleFunc("POST /api/users/login", userHandler.Login())
	router.HandleFunc("GET /api/articles/{slug}", articleHandler.GetArticle())

	router.HandleFunc("GET /api/profiles/{username}", profileHandler.GetProfile())

	// Protected routes
	authMiddleware := middleware.RequireAuth([]byte(cfg.JWT.SecretKey))
	router.HandleFunc("GET /api/user", authMiddleware(userHandler.GetCurrentUser()))
	router.HandleFunc("PUT /api/user", authMiddleware(userHandler.UpdateCurrentUser()))
	router.HandleFunc(
		"POST /api/profiles/{username}/follow",
		authMiddleware(profileHandler.Follow()),
	)
	router.HandleFunc(
		"DELETE /api/profiles/{username}/follow",
		authMiddleware(profileHandler.Unfollow()),
	)
	router.HandleFunc("POST /api/articles", authMiddleware(articleHandler.CreateArticle()))

	// Create HTTP server
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Handle graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Block until signal is received
	<-done
	log.Printf("Server stopping...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Printf("Server exited properly")
}
