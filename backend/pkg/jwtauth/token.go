package jwtauth

import (
	"errors"
	"fmt"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines our custom JWT claims structure.
type CustomClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// ValidateToken parses and validates a token string. It is shared across services.
func ValidateToken(tokenString string) (*jwt.Token, *CustomClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, nil, errors.New("JWT_SECRET environment variable is not set or empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, errors.New("invalid token")
}
