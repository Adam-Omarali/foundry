// utils/jwt.go
package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SessionTTL is how long a Foundry session lasts. The frontend stores the
// expiry with the session and treats the user as signed out after it.
const SessionTTL = 30 * 24 * time.Hour

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
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT issues a Foundry session token and returns it with its expiry.
func GenerateJWT(userID uuid.UUID, email string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(SessionTTL)
	claims := FoundryClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	return signed, expiresAt, err
}

// VerifyJWT checks the signature (HS256 only) and expiry. No database access.
func VerifyJWT(tokenStr string) (*FoundryClaims, error) {
	claims := &FoundryClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims,
		func(*jwt.Token) (interface{}, error) { return jwtSecret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if claims.UserID == uuid.Nil {
		return nil, fmt.Errorf("token has no user id")
	}
	return claims, nil
}
