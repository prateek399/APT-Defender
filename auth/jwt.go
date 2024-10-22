package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("your-secret-key"))
var Refresh_key = []byte("this-is-refresh-key")
var Access_key = []byte("this-is-access-key")

func GenerateToken(t *jwt.Token, secret_key []byte) (string, error) {
	return t.SignedString(secret_key)
}

// GenerateAccessToken generates a new access token based on the provided token string.
// It also returns the expiration time of the access token.
func GenerateAccessToken(refreshToken string) (string, int64, error) {
	// Parse the token string with the refresh key
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return Refresh_key, nil
	})
	if err != nil {
		return "", int64(0), err
	}
	if !token.Valid {
		return "", int64(0), nil
	}

	// Extract claims from the refresh token
	ref_claims := token.Claims.(*jwt.StandardClaims)

	// Set the expiration time and other claims for the new access token
	claims := jwt.StandardClaims{
		ExpiresAt: int64(time.Now().Add(time.Hour * 24 * 5).Unix()), // 5 days
		IssuedAt:  int64(time.Now().Unix()),
		Issuer:    ref_claims.Issuer,
	}

	// Create a new token with the claims and sign it with the access key
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := GenerateToken(newToken, Access_key)

	return accessToken, claims.ExpiresAt, err
}

// GenerateRefreshToken generates a refresh token for the given userAuth_id, username, and password.
func GenerateRefreshToken(username string) (string, error) {
	// Create the standard claims for the token
	claims := jwt.StandardClaims{
		ExpiresAt: int64(time.Now().Add(time.Hour * 24 * 30).Unix()), // 30 days
		IssuedAt:  int64(time.Now().Unix()),
		Issuer:    username,
	}
	// Create a new token with the standard claims using HMAC-SHA256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate the token using the Refresh_key
	return GenerateToken(token, Refresh_key)
}

// GetRefreshExp retrieves the expiration time of a refresh token
func GetRefreshExp(tokenString string) (int64, error) {
	// Parse the token with the refresh key and standard claims
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return Refresh_key, nil
	})

	// Return error if token parsing fails
	if err != nil {
		return int64(0), err
	}

	// Return error if token is not valid
	if !token.Valid {
		return int64(0), nil
	}

	// Retrieve the expiration time from the token claims
	claims := token.Claims.(*jwt.StandardClaims)
	return claims.ExpiresAt, nil
}
