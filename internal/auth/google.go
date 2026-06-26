package auth

import (
	"context"

	"folosy-backend/internal/domain"

	"google.golang.org/api/idtoken"
)

type GoogleVerifier struct {
	clientID string
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{clientID: clientID}
}

type GoogleIdentity struct {
	Sub           string
	Email         string
	EmailVerified bool
}

func (v *GoogleVerifier) Verify(ctx context.Context, idToken string) (GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return GoogleIdentity{}, domain.ErrInvalidGoogleToken
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return GoogleIdentity{}, domain.ErrInvalidGoogleToken
	}

	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok {
		return GoogleIdentity{}, domain.ErrInvalidGoogleToken
	}

	return GoogleIdentity{
		Sub:           payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}
