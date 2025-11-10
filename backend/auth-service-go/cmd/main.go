package main

import (
	"auth/internal/database"
	"auth/internal/handlers"
	"auth/internal/models"
	"lms/pkg/middleware"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// ... (Setup code remains the same) ...
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set")
	}

	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.StudentProfile{}, &models.InstructorProfile{}, &models.AdminProfile{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// --- ROUTER ---
	router := http.NewServeMux()

	// --- Public Routes ---
	router.HandleFunc("POST /login", handlers.Login(db))

	// --- Authenticated Routes (for any logged-in user) ---
	// Create a new router for routes that require any valid token
	authenticatedRoutes := http.NewServeMux()
	authenticatedRoutes.HandleFunc("GET /users/me", handlers.GetCurrentUser(db))
	// Apply the general auth middleware
	router.Handle("/users/", middleware.AuthMiddleware(authenticatedRoutes))

	// --- Admin-Only Routes ---
	adminRoutes := http.NewServeMux()
	adminRoutes.HandleFunc("POST /users", handlers.CreateUser(db))
	// Add the new GET and PUT routes for a specific user ID
	adminRoutes.HandleFunc("GET /users/{id}", handlers.GetUserByID(db))
	adminRoutes.HandleFunc("PUT /users/{id}", handlers.UpdateUser(db))

	// Apply the full security chain (Auth + Admin) to the admin routes
	protectedAdminRoutes := middleware.AuthMiddleware(middleware.AdminMiddleware(adminRoutes))
	router.Handle("/admin/", http.StripPrefix("/admin", protectedAdminRoutes))

	// this is for adding admin/troubleshooting via non-protected route.
	//router.Handle("/admin/", http.StripPrefix("/admin", adminRoutes))

	// --- START SERVER ---
	log.Println("🚀 Auth service starting on port 8081...")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
