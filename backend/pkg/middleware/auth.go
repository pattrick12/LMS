package middleware

import (
	"context"
	"lms/pkg/jwtauth"
	"net/http"
	"strings"
)

type ContextKey string

const UserRoleContextKey ContextKey = "userRole"
const UserIDContextKey ContextKey = "userID"

// ... (AuthMiddleware, AdminMiddleware, StudentMiddleware remain the same) ...
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		_, claims, err := jwtauth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserRoleContextKey, claims.Role)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleContextKey).(string)
		if !ok || role != "admin" {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func StudentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleContextKey).(string)
		if !ok {
			http.Error(w, "Forbidden: Role not found in context", http.StatusForbidden)
			return
		}
		if role != "student" && role != "ta" {
			http.Error(w, "Forbidden: Student or TA access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- NEW ---
// InstructorMiddleware checks if the user is an instructor.
// It must run *after* AuthMiddleware.
func InstructorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleContextKey).(string)
		if !ok {
			http.Error(w, "Forbidden: Role not found in context", http.StatusForbidden)
			return
		}
		if role != "instructor" {
			http.Error(w, "Forbidden: Instructor access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
