package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"folosy-backend/internal/auth"
	"folosy-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Register(ctx context.Context, user domain.User) (string, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByGoogleSub(ctx context.Context, sub string) (domain.User, error)
	LinkGoogleSub(ctx context.Context, userID, sub string) error
	CreateGoogleUser(ctx context.Context, email, username, sub string) (string, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type UserService struct {
	repo        UserRepository
	refreshRepo RefreshTokenRepository
	tokens      *auth.TokenService
	google      *auth.GoogleVerifier
}

func NewUserService(repo UserRepository, refreshRepo RefreshTokenRepository, tokens *auth.TokenService, google *auth.GoogleVerifier) *UserService {
	return &UserService{
		repo:        repo,
		refreshRepo: refreshRepo,
		tokens:      tokens,
		google:      google,
	}
}

// GetByID returns a user's profile by ID. Thin today, but it's the layer where
// future account-state checks will live.
func (s *UserService) GetByID(ctx context.Context, id string) (domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// LoginResult carries the two tokens a successful login produces.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

// issueTokens mints and stores a fresh access + refresh pair. Unexported so it
// stays callable only from the gated entry points (Login, Refresh, GoogleLogin).
func (s *UserService) issueTokens(ctx context.Context, userID string) (LoginResult, error) {
	accessToken, err := s.tokens.GenerateAccessToken(userID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue tokens: generate access token: %w", err)
	}

	refresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue tokens: generate refresh token: %w", err)
	}

	if err := s.refreshRepo.Create(ctx, userID, refresh.Hash, refresh.ExpiresAt); err != nil {
		return LoginResult{}, fmt.Errorf("issue tokens: store refresh token: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refresh.Raw,
	}, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return LoginResult{}, domain.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if !user.Password.Valid {
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return LoginResult{}, domain.ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("login: malformed stored hash for user %s: %w", user.ID, err)
	}

	return s.issueTokens(ctx, user.ID)
}

// GoogleLogin verifies a Google ID token and resolves it to a access + refresh tokens
func (s *UserService) GoogleLogin(ctx context.Context, idToken string) (LoginResult, error) {
	identity, err := s.google.Verify(ctx, idToken)
	if err != nil {
		return LoginResult{}, err
	}

	user, err := s.repo.GetByGoogleSub(ctx, identity.Sub)
	if err == nil {
		return s.issueTokens(ctx, user.ID)
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return LoginResult{}, fmt.Errorf("google login: lookup by sub: %w", err)
	}

	if !identity.EmailVerified {
		return LoginResult{}, domain.ErrInvalidGoogleToken
	}

	// Existing account with the same email — link Google onto it.
	user, err = s.repo.GetByEmail(ctx, identity.Email)
	if err == nil {
		if err := s.repo.LinkGoogleSub(ctx, user.ID, identity.Sub); err != nil {
			return LoginResult{}, fmt.Errorf("google login: link sub: %w", err)
		}
		return s.issueTokens(ctx, user.ID)
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return LoginResult{}, fmt.Errorf("google login: lookup by email: %w", err)
	}

	// 3. Brand-new user — username from the email's local part.
	username, _, _ := strings.Cut(identity.Email, "@")
	id, err := s.repo.CreateGoogleUser(ctx, identity.Email, username, identity.Sub)
	if err != nil {
		return LoginResult{}, fmt.Errorf("google login: create user: %w", err)
	}

	return s.issueTokens(ctx, id)
}

// Refresh redeems a raw refresh token for a brand-new access+refresh pair,
// rotating the token (single-use) and detecting reuse of an already-revoked one.
func (s *UserService) Refresh(ctx context.Context, rawToken string) (LoginResult, error) {
	stored, err := s.refreshRepo.GetByHash(ctx, s.tokens.HashRefreshToken(rawToken))
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			return LoginResult{}, domain.ErrInvalidRefreshToken
		}
		return LoginResult{}, fmt.Errorf("refresh: lookup token: %w", err)
	}

	if stored.RevokedAt.Valid {
		if err := s.refreshRepo.RevokeAllForUser(ctx, stored.UserID); err != nil {
			return LoginResult{}, fmt.Errorf("refresh: revoke reused token's family: %w", err)
		}
		return LoginResult{}, domain.ErrInvalidRefreshToken
	}

	if time.Now().After(stored.ExpiresAt) {
		return LoginResult{}, domain.ErrInvalidRefreshToken
	}

	if err := s.refreshRepo.Revoke(ctx, stored.ID); err != nil {
		return LoginResult{}, fmt.Errorf("refresh: revoke old token: %w", err)
	}

	return s.issueTokens(ctx, stored.UserID)
}

func (s *UserService) Register(ctx context.Context, email, username, password string) (domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		Email:    email,
		Username: username,
		// A registering user always has a password, so it's a valid (non-NULL) value.
		Password: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	id, err := s.repo.Register(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	user.ID = id
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return user, nil
}
