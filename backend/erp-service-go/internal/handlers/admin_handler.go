package handlers

import (
	"encoding/json"
	"erp/internal/models"
	"net/http"

	"gorm.io/gorm"
)

// AcademicRecordResponse is a custom struct to combine a student's full academic history.
type AcademicRecordResponse struct {
	CourseRegistrations  []models.Registration        `json:"courseRegistrations"`
	ProjectRegistrations []models.ProjectRegistration `json:"projectRegistrations"`
	AcademicStandings    []models.AcademicStanding    `json:"academicStandings"`
}

// AdminGetCourseRoster allows an admin to view the roster for any course.
func AdminGetCourseRoster(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseIDStr := r.PathValue("courseId")
		semester := r.PathValue("semester")

		// This is simpler than the instructor version because an admin has universal access
		// and does not need an ownership check.
		var registrations []models.Registration
		db.Where("course_id = ? AND semester = ?", courseIDStr, semester).Find(&registrations)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registrations)
	}
}

// GetStudentAcademicRecord allows an admin to view a student's complete academic history.
func GetStudentAcademicRecord(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		studentIDStr := r.PathValue("studentId")

		var courseRegs []models.Registration
		db.Where("user_id = ?", studentIDStr).Find(&courseRegs)

		var projectRegs []models.ProjectRegistration
		db.Where("user_id = ?", studentIDStr).Find(&projectRegs)

		var standings []models.AcademicStanding
		db.Where("user_id = ?", studentIDStr).Find(&standings)

		record := AcademicRecordResponse{
			CourseRegistrations:  courseRegs,
			ProjectRegistrations: projectRegs,
			AcademicStandings:    standings,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(record)
	}
}
