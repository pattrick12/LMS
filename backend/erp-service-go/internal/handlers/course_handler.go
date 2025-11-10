package handlers

import (
	"encoding/json"
	"erp/internal/models"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateCourseRequest now includes fields for all the new course rules.
type CreateCourseRequest struct {
	CourseCode      string      `json:"courseCode"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Credits         int         `json:"credits"`
	InstructorID    uuid.UUID   `json:"instructorId"`
	SemesterOffered string      `json:"semesterOffered"`
	CourseCap       *int        `json:"courseCap"`
	PrerequisiteIDs []uuid.UUID `json:"prerequisiteIds"` // We accept a list of Course IDs
}

// CreateCourse handles creating a new course with all its rules.
func CreateCourse(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateCourseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		course := models.Course{
			CourseCode:      req.CourseCode,
			Name:            req.Name,
			Description:     req.Description,
			Credits:         req.Credits,
			InstructorID:    req.InstructorID,
			SemesterOffered: req.SemesterOffered,
			CourseCap:       req.CourseCap,
		}

		tx := db.Begin()
		defer tx.Rollback()

		// First, create the course to get its ID
		if err := tx.Create(&course).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				http.Error(w, "Course code already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create course", http.StatusInternalServerError)
			return
		}

		// If there are prerequisites, find them and associate them.
		if len(req.PrerequisiteIDs) > 0 {
			var prereqs []models.Course
			if err := tx.Where("id IN ?", req.PrerequisiteIDs).Find(&prereqs).Error; err != nil {
				http.Error(w, "Failed to find prerequisites", http.StatusInternalServerError)
				return
			}
			if err := tx.Model(&course).Association("Prerequisites").Append(prereqs); err != nil {
				http.Error(w, "Failed to associate prerequisites", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			http.Error(w, "Failed to commit course creation", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(course)
	}
}

// ListCourses handles fetching all courses, including their prerequisites.
func ListCourses(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var courses []models.Course
		// Use Preload to efficiently fetch the related prerequisite courses.
		if err := db.Preload("Prerequisites").Find(&courses).Error; err != nil {
			log.Printf("ERROR: Failed to fetch courses: %v", err)
			http.Error(w, "Failed to retrieve course catalog", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(courses)
	}
}
