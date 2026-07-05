package middleware

import (
	"context"
	"net/http"
	"strings"

	"chatgoo/internal/pkg/jwt"
)

type contextKey string

const (
	// UserIDKey is the context key for the authenticated user's ID.
	UserIDKey contextKey = "userID"
	// UsernameKey is the context key for the authenticated user's username.
	UsernameKey contextKey = "username"
)

// publicPaths are routes that skip JWT authentication.
var publicPaths = map[string]bool{
	"/api/v1/auth/register": true,
	"/api/v1/auth/login":    true,
}

// Auth returns an HTTP middleware that validates JWT Bearer tokens.
// Public paths (register, login) are skipped.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"code":40101,"message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"code":40101,"message":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwt.Parse(parts[1], jwtSecret)
			if err != nil {
				http.Error(w, `{"code":40101,"message":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
