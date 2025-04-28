package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// JWTClaims custom claims type for JWT tokens
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID string) (string, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not set in environment")
	}

	expiryDuration, err := time.ParseDuration(os.Getenv("JWT_EXPIRY"))
	if err != nil {
		// Default to 24 hours if not specified or invalid
		expiryDuration = 24 * time.Hour
	}

	// Create the claims
	claims := JWTClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	return token.SignedString([]byte(jwtSecret))
}

// ValidateJWT validates a JWT token and returns the user ID if valid
func ValidateJWT(tokenString string) (string, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		return "", err
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not set in environment")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	// Validate the token and extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token")
} 