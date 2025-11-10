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

// GradeSubmission represents a single student's grade to be updated.
type GradeSubmission struct {
	UserID   uuid.UUID `json:"userId"`
	CourseID uuid.UUID `json:"courseId"`
	Semester string    `json:"semester"`
	Grade    string    `json:"grade"`
}

// SubmitGrades allows an instructor to submit final grades for multiple students.
func SubmitGrades(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instructorIDStr, _ := r.Context().Value(middleware.UserIDContextKey).(string)
		instructorID, _ := uuid.Parse(instructorIDStr)

		var submissions []GradeSubmission
		if err := json.NewDecoder(r.Body).Decode(&submissions); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if len(submissions) == 0 {
			http.Error(w, "No grades submitted", http.StatusBadRequest)
			return
		}

		tx := db.Begin()
		defer tx.Rollback()

		// BUSINESS LOGIC: Verify the instructor is assigned to the course they are submitting grades for.
		var course models.Course
		if err := tx.First(&course, "id = ?", submissions[0].CourseID).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if course.InstructorID != instructorID {
			http.Error(w, "Forbidden: You are not the instructor for this course", http.StatusForbidden)
			return
		}

		// Update grades for each student in the submission.
		for _, sub := range submissions {
			passFailStatus := "Pass" // Simple logic, can be expanded
			if sub.Grade == "F" {
				passFailStatus = "Fail"
			}

			result := tx.Model(&models.Registration{}).
				Where("user_id = ? AND course_id = ? AND semester = ?", sub.UserID, sub.CourseID, sub.Semester).
				Updates(map[string]interface{}{"grade": sub.Grade, "pass_fail_status": passFailStatus})

			if result.Error != nil || result.RowsAffected == 0 {
				log.Printf("WARN: Failed to update grade for user %s or registration not found", sub.UserID)
			}
		}

		tx.Commit()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Grades submitted successfully"}`))
	}
}

// GetCourseRoster allows an instructor to view all students registered for their course in a specific semester.
func GetCourseRoster(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		instructorIDStr, _ := r.Context().Value(middleware.UserIDContextKey).(string)
		courseIDStr := r.PathValue("courseId")
		semester := r.PathValue("semester")

		// BUSINESS LOGIC: Verify the instructor is assigned to this course.
		var course models.Course
		if err := db.First(&course, "id = ?", courseIDStr).Error; err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if course.InstructorID.String() != instructorIDStr {
			http.Error(w, "Forbidden: You are not the assigned instructor for this course", http.StatusForbidden)
			return
		}

		var registrations []models.Registration
		db.Where("course_id = ? AND semester = ?", courseIDStr, semester).Find(&registrations)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registrations)
	}
}
