// Package auth provides JWT generation/verification and password hashing.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims is the JWT payload.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// GenerateTokenPair creates a new signed access+refresh token pair.
func GenerateTokenPair(
	userID, email, accessSecret, refreshSecret string,
	accessTTL, refreshTTL time.Duration,
) (*TokenPair, error) {
	now := time.Now()

	accessClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			Issuer:    "ai-finance-tracker",
		},
		UserID: userID,
		Email:  email,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	atStr, err := at.SignedString([]byte(accessSecret))
	if err != nil {
		return nil, fmt.Errorf("auth: sign access token: %w", err)
	}

	refreshClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			Issuer:    "ai-finance-tracker",
		},
		UserID: userID,
		Email:  email,
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rtStr, err := rt.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("auth: sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:      atStr,
		RefreshToken:     rtStr,
		AccessExpiresAt:  now.Add(accessTTL),
		RefreshExpiresAt: now.Add(refreshTTL),
	}, nil
}

// VerifyToken parses and validates a JWT, returning the claims.
func VerifyToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("auth: invalid token")
	}
	return claims, nil
}

// HashPassword bcrypt-hashes a plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateSecureToken generates a cryptographically random hex token.
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
