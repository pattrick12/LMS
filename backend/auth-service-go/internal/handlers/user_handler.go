package handlers

import (
	"auth/internal/database"
	"auth/internal/models"
	"auth/internal/util"
	"encoding/json"
	"errors"
	"lms/pkg/middleware" // <-- THE FIX: Import the shared middleware package
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ... (GetUserByID, UpdateUser, CreateUser functions remain the same) ...

// GetCurrentUser handles a logged-in user fetching their own profile.
func GetCurrentUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// --- THE FIX ---
		// We now use the shared, consistent key from the middleware package
		// to retrieve the userID from the request context.
		userIDStr, ok := r.Context().Value(middleware.UserIDContextKey).(string)
		if !ok {
			util.WriteJSON(w, http.StatusUnauthorized, util.H{"error": "Could not identify user from context"})
			return
		}

		var user models.User
		err := db.Preload("StudentProfile").Preload("InstructorProfile").Preload("AdminProfile").First(&user, "id = ?", userIDStr).Error
		if err != nil {
			util.WriteJSON(w, http.StatusNotFound, util.H{"error": "User not found"})
			return
		}

		util.WriteJSON(w, http.StatusOK, util.H{"user": user})
	}
}

func GetUserByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.PathValue("id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			util.WriteJSON(w, http.StatusBadRequest, util.H{"error": "Invalid user ID format"})
			return
		}

		var user models.User
		err = db.Preload("StudentProfile").Preload("InstructorProfile").Preload("AdminProfile").First(&user, "id = ?", userID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				util.WriteJSON(w, http.StatusNotFound, util.H{"error": "User not found"})
				return
			}
			log.Printf("ERROR: Failed to fetch user: %v", err)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to retrieve user"})
			return
		}

		util.WriteJSON(w, http.StatusOK, util.H{"user": user})
	}
}

func UpdateUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.PathValue("id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			util.WriteJSON(w, http.StatusBadRequest, util.H{"error": "Invalid user ID format"})
			return
		}

		var user models.User
		if err := db.First(&user, "id = ?", userID).Error; err != nil {
			util.WriteJSON(w, http.StatusNotFound, util.H{"error": "User not found"})
			return
		}

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			util.WriteJSON(w, http.StatusBadRequest, util.H{"error": "Invalid request payload"})
			return
		}

		tx := db.Begin()
		defer tx.Rollback()

		var updateErr error
		switch user.Role {
		case "student", "ta":
			updateErr = tx.Model(&models.StudentProfile{}).Where("user_id = ?", userID).Updates(updates).Error
		case "instructor":
			updateErr = tx.Model(&models.InstructorProfile{}).Where("user_id = ?", userID).Updates(updates).Error
		case "admin":
			updateErr = tx.Model(&models.AdminProfile{}).Where("user_id = ?", userID).Updates(updates).Error
		}

		if updateErr != nil {
			log.Printf("ERROR: Failed to update profile: %v", updateErr)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to update user profile"})
			return
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("ERROR: Failed to commit transaction: %v", err)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to update user"})
			return
		}

		util.WriteJSON(w, http.StatusOK, util.H{"message": "User updated successfully"})
	}
}

type CreateUserRequest struct {
	Email             string                    `json:"email"`
	Password          string                    `json:"password"`
	Role              string                    `json:"role"`
	StudentProfile    *models.StudentProfile    `json:"studentProfile,omitempty"`
	InstructorProfile *models.InstructorProfile `json:"instructorProfile,omitempty"`
	AdminProfile      *models.AdminProfile      `json:"adminProfile,omitempty"`
}

func CreateUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteJSON(w, http.StatusBadRequest, util.H{"error": "Invalid request payload"})
			return
		}

		if req.Email == "" || req.Password == "" || req.Role == "" {
			util.WriteJSON(w, http.StatusBadRequest, util.H{"error": "Email, password, and role are required"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("ERROR: could not hash password: %v", err)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to process request"})
			return
		}

		tx := db.Begin()
		if tx.Error != nil {
			log.Printf("ERROR: could not begin transaction: %v", err)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to create user"})
			return
		}
		defer tx.Rollback()

		user := models.User{
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			Role:         req.Role,
		}
		if err := tx.Create(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				util.WriteJSON(w, http.StatusConflict, util.H{"error": "Email already exists"})
			} else {
				log.Printf("ERROR: Database error when inserting user: %v", err)
				util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to create user"})
			}
			return
		}

		var profileData interface{}
		switch req.Role {
		case "student", "ta":
			profileData = req.StudentProfile
		case "instructor":
			profileData = req.InstructorProfile
		case "admin":
			profileData = req.AdminProfile
		}

		err = database.CreateUserProfile(tx, user.ID, req.Role, profileData)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "roll_no") {
				util.WriteJSON(w, http.StatusConflict, util.H{"error": "Roll number already exists"})
			} else if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "employee_id") {
				util.WriteJSON(w, http.StatusConflict, util.H{"error": "Employee ID already exists"})
			} else {
				util.WriteJSON(w, http.StatusBadRequest, util.H{"error": err.Error()})
			}
			return
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("ERROR: could not commit transaction: %v", err)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to create user"})
			return
		}

		util.WriteJSON(w, http.StatusCreated, util.H{"message": "User created successfully", "userID": user.ID})
	}
}
