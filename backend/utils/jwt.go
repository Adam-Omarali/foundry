// utils/jwt.go
package utils

import (
	"fmt"
	"os"
	"time"

	"foundry/backend/db"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// InitializeJWT must be called after environment variables are loaded
func InitializeJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is not set")
	}
	jwtSecret = []byte(secret)
	return nil
}

type FoundryClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// Generate JWT with 24h expiration
func GenerateJWT(email string) (string, error) {
	claims := FoundryClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Verify JWT
func VerifyJWT(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &FoundryClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if claims, ok := token.Claims.(*FoundryClaims); ok && token.Valid {
		userId := db.SignInUser(claims.Email)

		// Get user information including plan
		user := db.GetUserByEmail(claims.Email)

		ret := map[string]interface{}{
			"user_id": userId,
			"email":   claims.Email,
			"plan":    user.Plan,
		}
		return ret, nil
	} else {
		return nil, err
	}
}
