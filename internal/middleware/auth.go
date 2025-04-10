package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// userIDContextKey is the context key for the user ID
const userIDContextKey = contextKey("userID")

// GenericErrorModel represents the API error response body
type GenericErrorModel struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

// RequireAuth middleware validates the JWT token and adds the user ID to the request context
func RequireAuth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Token ") {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Token ")
			if tokenString == "" {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			claims, ok := token.Claims.(*jwt.StandardClaims)
			if !ok || claims.Subject == "" {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			if claims.ExpiresAt < time.Now().Unix() {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, []string{"Unauthorized"})
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

// respondWithError sends an error response with the given status code and errors
func respondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	json.NewEncoder(w).Encode(response)
}
