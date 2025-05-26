package database

import (
	"database/sql"
	"log"

	"github.com/Nilesh2000/conduit/internal/config"
)

// Setup initializes and returns a database connection with the given configuration
func Setup(cfg *config.Config) *sql.DB {
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure database connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// Ping database to check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}
