package handlers

import (
	"auth/internal/models"
	"auth/internal/oauth"
	"auth/internal/util"
	"encoding/json"
	"errors"
	"log" // <-- Make sure log is imported
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginRequest defines the expected JSON payload for a login attempt.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles user authentication and issues a JWT upon success.
func Login(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteJSON(w, http.StatusBadRequest, util.H{"error": "Invalid request payload"})
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				util.WriteJSON(w, http.StatusUnauthorized, util.H{"error": "Invalid credentials"})
				return
			}
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to login"})
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			util.WriteJSON(w, http.StatusUnauthorized, util.H{"error": "Invalid credentials"})
			return
		}

		// If login is successful, generate a JWT
		token, err := oauth.GenerateToken(user.ID.String(), user.Role)
		if err != nil {
			// --- THIS IS THE NEW LOGGING ---
			// Print the specific, detailed error to the server console for debugging.
			log.Printf("ERROR: Could not generate JWT: %v", err)
			util.WriteJSON(w, http.StatusInternalServerError, util.H{"error": "Failed to generate token"})
			return
		}

		util.WriteJSON(w, http.StatusOK, util.H{"message": "Login successful", "token": token})
	}
}
