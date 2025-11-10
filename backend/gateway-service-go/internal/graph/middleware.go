package graph

import (
	"context"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// authTokenKey is the key we use to store the Authorization header in the context.
const authTokenKey contextKey = "authToken"

// AuthMiddleware is a simple middleware for the gateway to extract the token header.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the full header (e.g., "Bearer <token>")
		authHeader := r.Header.Get("Authorization")
		// Put it into the context for the resolvers to use.
		ctx := context.WithValue(r.Context(), authTokenKey, authHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
