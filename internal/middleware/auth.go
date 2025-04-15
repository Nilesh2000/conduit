package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// UserIDContextKey is the context key for the user ID
const UserIDContextKey = contextKey("userID")

// GenericErrorModel represents the API error response body
type GenericErrorModel struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

// RequireAuth middleware validates the JWT token and adds the user ID to the request context
func RequireAuth(jwtSecret []byte) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the request context
			ctx := r.Context()
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Token ") {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Extract the token from the Authorization header
			tokenString := strings.TrimPrefix(authHeader, "Token ")
			if tokenString == "" {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Parse the token
			token, err := jwt.ParseWithClaims(
				tokenString,
				&jwt.StandardClaims{},
				func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				},
			)
			if err != nil || !token.Valid {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Extract the claims from the token
			claims, ok := token.Claims.(*jwt.StandardClaims)
			if !ok || claims.Subject == "" {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Check if the token has expired
			if claims.ExpiresAt < time.Now().Unix() {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Parse the user ID from the claims
			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Add the user ID to the request context
			ctx = context.WithValue(ctx, UserIDContextKey, userID)

			// Serve the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int64)
	return userID, ok
}

// respondWithError sends an error response with the given status code and errors
func respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
