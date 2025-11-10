package database

import (
	"auth/internal/models"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUserProfile is a helper function to create the correct profile based on role.
// It now accepts the specific profile struct as an interface{}.
func CreateUserProfile(tx *gorm.DB, userID uuid.UUID, role string, profileData interface{}) error {
	switch role {
	case "student", "ta":
		profile, ok := profileData.(*models.StudentProfile)
		if !ok || profile == nil {
			return errors.New("student profile data is required")
		}
		if profile.RollNo == "" {
			return errors.New("roll number is required for students")
		}
		profile.UserID = userID
		if err := tx.Create(profile).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				// Return a specific, structured error
				return fmt.Errorf("duplicate key error: roll number '%s' already exists", profile.RollNo)
			}
			return fmt.Errorf("failed to create student profile: %w", err)
		}
	case "instructor":
		profile, ok := profileData.(*models.InstructorProfile)
		if !ok || profile == nil {
			return errors.New("instructor profile data is required")
		}
		if profile.EmployeeID == "" {
			return errors.New("employee ID is required for instructors")
		}
		profile.UserID = userID
		if err := tx.Create(profile).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("duplicate key error: employee ID '%s' already exists", profile.EmployeeID)
			}
			return fmt.Errorf("failed to create instructor profile: %w", err)
		}
	case "admin":
		profile, ok := profileData.(*models.AdminProfile)
		if !ok || profile == nil {
			return errors.New("admin profile data is required")
		}
		if profile.EmployeeID == "" {
			return errors.New("employee ID is required for admins")
		}
		profile.UserID = userID
		if err := tx.Create(profile).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("duplicate key error: employee ID '%s' already exists", profile.EmployeeID)
			}
			return fmt.Errorf("failed to create admin profile: %w", err)
		}
	default:
		return fmt.Errorf("invalid or unsupported role: %s", role)
	}
	return nil
}
