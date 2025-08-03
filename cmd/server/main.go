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

	"github.com/Nilesh2000/conduit/internal/config"
	"github.com/Nilesh2000/conduit/internal/handler"
	"github.com/Nilesh2000/conduit/internal/middleware"
	"github.com/Nilesh2000/conduit/internal/repository/postgres"
	"github.com/Nilesh2000/conduit/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup database
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Configure database connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// Ping database to check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize repositories
	userRepository := postgres.NewUserRepository(db)
	profileRepository := postgres.NewProfileRepository(db)
	articleRepository := postgres.NewArticleRepository(db)
	tagRepository := postgres.NewTagRepository(db)
	commentRepository := postgres.NewCommentRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepository, cfg.JWT.SecretKey, cfg.JWT.Expiry)
	profileService := service.NewProfileService(userRepository, profileRepository)
	articleService := service.NewArticleService(articleRepository, profileRepository)
	tagService := service.NewTagService(tagRepository)
	commentService := service.NewCommentService(commentRepository, articleRepository)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	profileHandler := handler.NewProfileHandler(profileService)
	articleHandler := handler.NewArticleHandler(articleService)
	tagHandler := handler.NewTagHandler(tagService)
	commentHandler := handler.NewCommentHandler(commentService)
	healthHandler := handler.NewHealthHandler(cfg.Version)

	// Initialize middleware
	authMiddleware := middleware.RequireAuth([]byte(cfg.JWT.SecretKey))

	// Setup router
	router := http.NewServeMux()

	// Apply middleware
	handler := middleware.LoggingMiddleware(router)

	// Health endpoint
	router.HandleFunc("GET /health", healthHandler.Health())

	// Article routes
	router.HandleFunc("GET /api/articles", articleHandler.ListArticles())
	router.HandleFunc("GET /api/articles/feed", authMiddleware(articleHandler.GetArticlesFeed()))
	router.HandleFunc("POST /api/articles", authMiddleware(articleHandler.CreateArticle()))
	router.HandleFunc("GET /api/articles/{slug}", articleHandler.GetArticle())
	router.HandleFunc("PUT /api/articles/{slug}", authMiddleware(articleHandler.UpdateArticle()))
	router.HandleFunc("DELETE /api/articles/{slug}", authMiddleware(articleHandler.DeleteArticle()))

	// Comment routes
	router.HandleFunc("GET /api/articles/{slug}/comments", commentHandler.GetComments())
	router.HandleFunc(
		"POST /api/articles/{slug}/comments",
		authMiddleware(commentHandler.CreateComment()),
	)
	router.HandleFunc(
		"DELETE /api/articles/{slug}/comments/{id}",
		authMiddleware(commentHandler.DeleteComment()),
	)

	// Favorite routes
	router.HandleFunc(
		"POST /api/articles/{slug}/favorite",
		authMiddleware(articleHandler.FavoriteArticle()),
	)
	router.HandleFunc(
		"DELETE /api/articles/{slug}/favorite",
		authMiddleware(articleHandler.UnfavoriteArticle()),
	)

	// Profile routes
	router.HandleFunc("GET /api/profiles/{username}", profileHandler.GetProfile())
	router.HandleFunc(
		"POST /api/profiles/{username}/follow",
		authMiddleware(profileHandler.Follow()),
	)
	router.HandleFunc(
		"DELETE /api/profiles/{username}/follow",
		authMiddleware(profileHandler.Unfollow()),
	)

	// Tag routes
	router.HandleFunc("GET /api/tags", tagHandler.GetTags())

	// User and Authentication routes
	router.HandleFunc("POST /api/users/login", userHandler.Login())
	router.HandleFunc("POST /api/users", userHandler.Register())
	router.HandleFunc("GET /api/user", authMiddleware(userHandler.GetCurrentUser()))
	router.HandleFunc("PUT /api/user", authMiddleware(userHandler.UpdateCurrentUser()))

	// Create HTTP server
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
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
