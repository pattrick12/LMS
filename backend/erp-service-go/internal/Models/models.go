package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Course model with prerequisites, semester, and capping.
type Course struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CourseCode      string    `gorm:"type:varchar(20);uniqueIndex;not null"`
	Name            string    `gorm:"type:varchar(255);not null"`
	Description     string    `gorm:"type:text"`
	Credits         int       `gorm:"not null"`
	InstructorID    uuid.UUID `gorm:"type:uuid;not null"`
	SemesterOffered string    `gorm:"type:varchar(50)"`
	CourseCap       *int
	Prerequisites   []*Course `gorm:"many2many:course_prerequisites;"`
	AntiRequisites  []*Course `gorm:"many2many:course_anti_requisites;"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Registration correctly stores semester-wise grades and status.
type Registration struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CourseID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Semester       string    `gorm:"type:varchar(50);primaryKey"` // e.g., "Monsoon 2025"
	Grade          *string   `gorm:"type:varchar(10)"`
	PassFailStatus *string   `gorm:"type:varchar(20)"`
	CreatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// ProjectRegistration stores academic project details, including grades.
type ProjectRegistration struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	InstructorID uuid.UUID `gorm:"type:uuid;not null"`
	Semester     string    `gorm:"type:varchar(50);not null"`
	ProjectTitle string    `gorm:"type:varchar(255);not null"`
	Description  string    `gorm:"type:text"`
	Credits      int       `gorm:"not null"`
	Grade        *string   `gorm:"type:varchar(10)"`
	Status       string    `gorm:"type:varchar(50);default:'In Progress'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AcademicStanding tracks a student's status and progress for a specific semester.
type AcademicStanding struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Semester       string    `gorm:"type:varchar(50);primaryKey"` // e.g., "Monsoon 2025"
	SemesterNumber int       `gorm:"not null"`                    // e.g., 7
	Status         string    `gorm:"type:varchar(50);not null"`   // e.g., 'Enrolled', 'On Break'
	SGPA           *float64
	CGPA           *float64
}
