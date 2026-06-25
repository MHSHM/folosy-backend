package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"folosy-backend/internal/auth"
	"folosy-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Register(ctx context.Context, user domain.User) (string, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
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
}

func NewUserService(repo UserRepository, refreshRepo RefreshTokenRepository, tokens *auth.TokenService) *UserService {
	return &UserService{
		repo:        repo,
		refreshRepo: refreshRepo,
		tokens:      tokens,
	}
}

// LoginResult carries the two tokens a successful login produces.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *UserService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return LoginResult{}, domain.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return LoginResult{}, domain.ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("login: malformed stored hash for user %s: %w", user.ID, err)
	}

	// short-lived access JWT (15 minutes)
	accessToken, err := s.tokens.GenerateAccessToken(user.ID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("login: generate access token: %w", err)
	}

	// long-lived refresh token (7 days)
	refresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("login: generate refresh token: %w", err)
	}

	if err := s.refreshRepo.Create(ctx, user.ID, refresh.Hash, refresh.ExpiresAt); err != nil {
		return LoginResult{}, fmt.Errorf("login: store refresh token: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refresh.Raw,
	}, nil
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

	accessToken, err := s.tokens.GenerateAccessToken(stored.UserID)
	if err != nil {
		return LoginResult{}, fmt.Errorf("refresh: generate access token: %w", err)
	}

	newRefresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("refresh: generate refresh token: %w", err)
	}

	if err := s.refreshRepo.Create(ctx, stored.UserID, newRefresh.Hash, newRefresh.ExpiresAt); err != nil {
		return LoginResult{}, fmt.Errorf("refresh: store new refresh token: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefresh.Raw,
	}, nil
}

func (s *UserService) Register(ctx context.Context, email, username, password string) (domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
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
