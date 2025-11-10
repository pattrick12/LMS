package handlers

import (
	"encoding/json"
	"erp/internal/models"
	"lms/pkg/middleware"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	CourseID uuid.UUID `json:"courseId"`
	Semester string    `json:"semester"`
}

// RegisterForCourse now contains business logic for validation.
func RegisterForCourse(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr, _ := r.Context().Value(middleware.UserIDContextKey).(string)
		userID, _ := uuid.Parse(userIDStr)

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		tx := db.Begin()
		defer tx.Rollback()

		// 1. Fetch the course and its prerequisites
		var course models.Course
		if err := tx.Preload("Prerequisites").First(&course, "id = ?", req.CourseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		// 2. BUSINESS LOGIC: Check Course Cap
		if course.CourseCap != nil {
			var currentRegistrations int64
			tx.Model(&models.Registration{}).Where("course_id = ? AND semester = ?", req.CourseID, req.Semester).Count(&currentRegistrations)
			if currentRegistrations >= int64(*course.CourseCap) {
				http.Error(w, "Course registration is full (cap reached)", http.StatusConflict)
				return
			}
		}

		// 3. BUSINESS LOGIC: Check Prerequisites
		if len(course.Prerequisites) > 0 {
			var completedPrereqs int64
			prereqIDs := []uuid.UUID{}
			for _, p := range course.Prerequisites {
				prereqIDs = append(prereqIDs, p.ID)
			}
			// Find how many of the required prerequisite courses the student has passed.
			// NOTE: This assumes a simple 'Pass' status. A real system might check specific grades.
			tx.Model(&models.Registration{}).
				Where("user_id = ? AND course_id IN ? AND pass_fail_status = ?", userID, prereqIDs, "Pass").
				Count(&completedPrereqs)

			if completedPrereqs < int64(len(course.Prerequisites)) {
				http.Error(w, "Prerequisite courses not met", http.StatusForbidden)
				return
			}
		}

		// 4. If all checks pass, create the registration
		registration := models.Registration{
			UserID:   userID,
			CourseID: req.CourseID,
			Semester: req.Semester,
		}

		if err := tx.Create(&registration).Error; err != nil {
			log.Printf("ERROR: Failed to create registration: %v", err)
			http.Error(w, "Failed to register for course. You may already be registered for it this semester.", http.StatusConflict)
			return
		}

		tx.Commit()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"Successfully registered for course"}`))
	}
}

// ... (ListMyRegistrations and DropCourse remain the same) ...
func ListMyRegistrations(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr, _ := r.Context().Value(middleware.UserIDContextKey).(string)
		var registrations []models.Registration
		db.Where("user_id = ?", userIDStr).Find(&registrations)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registrations)
	}
}

func DropCourse(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr, _ := r.Context().Value(middleware.UserIDContextKey).(string)
		courseIDStr := r.PathValue("courseId")
		semester := r.PathValue("semester")
		err := db.Delete(&models.Registration{
			UserID:   uuid.MustParse(userIDStr),
			CourseID: uuid.MustParse(courseIDStr),
			Semester: semester,
		}).Error
		if err != nil {
			log.Printf("ERROR: Failed to drop course: %v", err)
			http.Error(w, "Failed to drop course", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Successfully dropped course"}`))
	}
}
