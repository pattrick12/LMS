package oauth

import (
	"errors"
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines our custom JWT claims structure. It embeds the standard
// RegisteredClaims and adds our own custom 'Role' field.
type CustomClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT for a given user ID and role.
func GenerateToken(userID, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set or empty")
	}

	// Create our custom claims
	claims := CustomClaims{
		role, // Set our custom role claim
		jwt.RegisteredClaims{
			Issuer:    "lms-auth-service",
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create a new token object, specifying the signing method and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key to get the complete, signed token string.
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return tokenString, nil
}
