package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nilesh2000/conduit/internal/config"
	"github.com/Nilesh2000/conduit/internal/database"
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
	db := database.Setup(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Initialize repositories
	userRepository := postgres.NewUserRepository(db)
	profileRepository := postgres.NewProfileRepository(db)
	articleRepository := postgres.NewArticleRepository(db)
	tagRepository := postgres.NewTagRepository(db)
	commentRepository := postgres.NewCommentRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepository, cfg.JWT.SecretKey, cfg.JWT.Expiry)
	profileService := service.NewProfileService(profileRepository)
	articleService := service.NewArticleService(articleRepository, profileRepository)
	tagService := service.NewTagService(tagRepository)
	commentService := service.NewCommentService(commentRepository, articleRepository)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	profileHandler := handler.NewProfileHandler(profileService)
	articleHandler := handler.NewArticleHandler(articleService)
	tagHandler := handler.NewTagHandler(tagService)
	commentHandler := handler.NewCommentHandler(commentService)

	// Setup router
	router := http.NewServeMux()

	// Apply middleware
	handler := middleware.LoggingMiddleware(router)

	// Public routes

	// Auth routes
	router.HandleFunc("POST /api/users", userHandler.Register())
	router.HandleFunc("POST /api/users/login", userHandler.Login())

	// Profile routes
	router.HandleFunc("GET /api/profiles/{username}", profileHandler.GetProfile())

	// Article routes
	router.HandleFunc("GET /api/articles/{slug}", articleHandler.GetArticle())

	// Comment routes
	router.HandleFunc("GET /api/articles/{slug}/comments", commentHandler.GetComments())

	// Tag routes
	router.HandleFunc("GET /api/tags", tagHandler.GetTags())

	// Protected routes
	authMiddleware := middleware.RequireAuth([]byte(cfg.JWT.SecretKey))

	// User routes
	router.HandleFunc("GET /api/user", authMiddleware(userHandler.GetCurrentUser()))
	router.HandleFunc("PUT /api/user", authMiddleware(userHandler.UpdateCurrentUser()))

	// Profile routes
	router.HandleFunc(
		"POST /api/profiles/{username}/follow",
		authMiddleware(profileHandler.Follow()),
	)
	router.HandleFunc(
		"DELETE /api/profiles/{username}/follow",
		authMiddleware(profileHandler.Unfollow()),
	)

	// Article routes
	router.HandleFunc("POST /api/articles", authMiddleware(articleHandler.CreateArticle()))
	router.HandleFunc("PUT /api/articles/{slug}", authMiddleware(articleHandler.UpdateArticle()))
	router.HandleFunc("DELETE /api/articles/{slug}", authMiddleware(articleHandler.DeleteArticle()))

	// Favorite routes
	router.HandleFunc(
		"POST /api/articles/{slug}/favorite",
		authMiddleware(articleHandler.FavoriteArticle()),
	)
	router.HandleFunc(
		"DELETE /api/articles/{slug}/favorite",
		authMiddleware(articleHandler.UnfavoriteArticle()),
	)

	// Comment routes
	router.HandleFunc(
		"POST /api/articles/{slug}/comments",
		authMiddleware(commentHandler.CreateComment()),
	)
	router.HandleFunc(
		"DELETE /api/articles/{slug}/comments/{id}",
		authMiddleware(commentHandler.DeleteComment()),
	)

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
