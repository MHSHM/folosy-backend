package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService creates and verifies tokens. It holds the signing secret and
// the access-token lifetime so the rest of the app can ask for a token without
// knowing anything about how tokens work internally.
type TokenService struct {
	secret     []byte        // the signing key; only this server knows it
	accessTTL  time.Duration // how long an access token stays valid
	refreshTTL time.Duration // how long a refresh token stays valid
}

func NewTokenService(secret string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type RefreshToken struct {
	Raw       string    // the unguessable token — sent to the client, never stored
	Hash      string    // SHA-256 of Raw — stored in the DB, never sent
	ExpiresAt time.Time // when it expires (now + refreshTTL)
}

// GenerateAccessToken returns a signed JWT whose subject is the user's ID.
func (s *TokenService) GenerateAccessToken(userID string) (string, error) {
	now := time.Now()

	// the claims — the data embedded in the token's payload.
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
	}

	// assemble header (HS256) + claims into an unsigned token object.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign with the secret, producing the final header.payload.signature string.
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}

// VerifyAccessToken checks that a token string is genuine and unexpired, and
// returns the user ID it carries.
func (s *TokenService) VerifyAccessToken(tokenString string) (string, error) {
	// ParseWithClaims re-computes the signature and compares it, and validates
	// exp for us. It calls the keyfunc below to ask "which key do I verify with?"
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secret, nil
		})
	if err != nil {
		return "", fmt.Errorf("verify access token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("verify access token: invalid claims")
	}

	return claims.Subject, nil
}

// GenerateRefreshToken creates a new, unguessable refresh token
func (s *TokenService) GenerateRefreshToken() (RefreshToken, error) {
	// cryptographically-secure random bytes, encoded to a string.
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return RefreshToken{}, fmt.Errorf("read random bytes: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(randomBytes)

	return RefreshToken{
		Raw:       raw,
		Hash:      s.HashRefreshToken(raw),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}, nil
}

// HashRefreshToken returns the SHA-256 (hex) of a raw refresh token.
func (s *TokenService) HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
