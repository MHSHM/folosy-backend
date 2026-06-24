package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService creates and verifies tokens. It holds the signing secret and
// the access-token lifetime so the rest of the app can ask for a token without
// knowing anything about how tokens work internally.
type TokenService struct {
	secret    []byte        // the signing key; only this server knows it
	accessTTL time.Duration // how long an access token stays valid
}

func NewTokenService(secret string, accessTTL time.Duration) *TokenService {
	return &TokenService{
		secret:    []byte(secret),
		accessTTL: accessTTL,
	}
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
