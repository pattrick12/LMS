package main

import (
	"erp/internal/database"
	"erp/internal/handlers"
	"erp/internal/models"
	"lms/pkg/middleware"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("[ERP Service] Could not connect to the database: %v", err)
	}

	err = db.AutoMigrate(
		&models.Course{},
		&models.Registration{},
		&models.ProjectRegistration{},
		&models.AcademicStanding{},
	)
	if err != nil {
		log.Fatalf("[ERP Service] Failed to migrate database: %v", err)
	}

	router := http.NewServeMux()
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "erp service is up and running"}`))
	})

	// --- General Course Routes ---
	courseRouter := http.NewServeMux()
	courseRouter.Handle("POST /", middleware.AdminMiddleware(http.HandlerFunc(handlers.CreateCourse(db))))
	courseRouter.Handle("GET /", http.HandlerFunc(handlers.ListCourses(db))) // Publicly viewable
	courseRouter.Handle("GET /{courseId}/roster/{semester}", middleware.InstructorMiddleware(http.HandlerFunc(handlers.GetCourseRoster(db))))
	router.Handle("/courses/", http.StripPrefix("/courses", middleware.AuthMiddleware(courseRouter)))
	router.Handle("/courses", middleware.AuthMiddleware(courseRouter))

	// --- Student Registration Routes ---
	regRouter := http.NewServeMux()
	regRouter.Handle("POST /", middleware.StudentMiddleware(http.HandlerFunc(handlers.RegisterForCourse(db))))
	regRouter.Handle("GET /me", middleware.StudentMiddleware(http.HandlerFunc(handlers.ListMyRegistrations(db))))
	regRouter.Handle("DELETE /{courseId}/{semester}", middleware.StudentMiddleware(http.HandlerFunc(handlers.DropCourse(db))))
	regRouter.Handle("POST /grades", middleware.InstructorMiddleware(http.HandlerFunc(handlers.SubmitGrades(db))))
	router.Handle("/registrations/", http.StripPrefix("/registrations", middleware.AuthMiddleware(regRouter)))
	router.Handle("/registrations", middleware.AuthMiddleware(regRouter))

	// --- NEW: Admin-specific ERP Routes ---
	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("GET /roster/{courseId}/{semester}", handlers.AdminGetCourseRoster(db))
	adminRouter.HandleFunc("GET /records/student/{studentId}", handlers.GetStudentAcademicRecord(db))
	// All routes in this group are protected by both Auth and Admin middleware
	router.Handle("/admin/erp/", http.StripPrefix("/admin/erp", middleware.AuthMiddleware(middleware.AdminMiddleware(adminRouter))))

	log.Println("🚀 ERP service starting on port 8082...")
	if err := http.ListenAndServe(":8082", router); err != nil {
		log.Fatalf("[ERP Service] Could not start server: %s\n", err)
	}
}
