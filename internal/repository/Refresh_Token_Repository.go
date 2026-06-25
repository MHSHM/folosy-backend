package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"folosy-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

// GetByHash looks a token up by its SHA-256 hash.
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var rt domain.RefreshToken
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&rt.ID, &rt.UserID, &rt.ExpiresAt, &rt.RevokedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RefreshToken{}, domain.ErrRefreshTokenNotFound
		}
		return domain.RefreshToken{}, fmt.Errorf("get refresh token by hash: %w", err)
	}

	return rt, nil
}

// Revoke soft-revokes a single token by id
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllForUser revokes every still-active token for a user — the response to
// detected reuse (log out everywhere).
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens for user: %w", err)
	}

	return nil
}

// DeleteExpired hard-deletes every expired row (revoked or not — once expired a
// token is unusable, so its row carries no further value). Returns the number of
// rows removed so the cleanup job can log it.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < now()`
	tag, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}

	return tag.RowsAffected(), nil
}
