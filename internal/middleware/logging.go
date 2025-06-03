package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware is a middleware that logs all requests and their responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the request details
		duration := time.Since(start)
		log.Printf(
			"Method: %s, Path: %s, Status: %d, Duration: %v",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			duration,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}
