package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Nilesh2000/conduit/internal/response"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// UserIDContextKey is the context key for the user ID
const UserIDContextKey = contextKey("userID")

// RequireAuth middleware validates the JWT token and adds the user ID to the request context
func RequireAuth(jwtSecret []byte) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the request context
			ctx := r.Context()

			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Token ") {
				response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Extract the token from the Authorization header
			tokenString := strings.TrimPrefix(authHeader, "Token ")
			if tokenString == "" {
				response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Parse the token
			token, err := jwt.ParseWithClaims(
				tokenString,
				&jwt.RegisteredClaims{},
				func(token *jwt.Token) (any, error) {
					return jwtSecret, nil
				},
			)
			if err != nil || !token.Valid {
				response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Extract the claims from the token
			claims, ok := token.Claims.(*jwt.RegisteredClaims)
			if !ok || claims.Subject == "" {
				response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Check if the token has expired
			expTime, err := claims.GetExpirationTime()
			if err != nil || expTime.Before(time.Now()) {
				response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			// Parse the user ID from the claims
			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				response.RespondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
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
