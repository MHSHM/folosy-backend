package repository

import (
	"context"
	"errors"
	"fmt"

	"folosy-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Register(ctx context.Context, user domain.User) (string, error) {
	query := `
		INSERT INTO users (email, username, password)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, user.Email, user.Username, user.Password).Scan(&id)
	if err != nil {
		// Postgres rejects a duplicate email with SQLSTATE 23505 (unique
		// violation). email is the only unique column on users, so any 23505
		// here means the email is taken. errors.As pulls the typed *pgconn.PgError
		// out of the wrapped chain so we can inspect its SQLSTATE code.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", domain.ErrEmailExists
		}

		return "", fmt.Errorf("register new user: %w", err)
	}

	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	query := `
		SELECT id, email, username
		FROM users
		WHERE id = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT id, email, username, password
		FROM users
		WHERE email = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	if err != nil {
		// pgx.ErrNoRows means the WHERE matched nothing — no user with that
		// email. Translate that database-specific signal into our domain sentinel
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// GetByGoogleSub looks a user up by their Google subject ID — the "is this a
// returning Google user?" check.
func (r *UserRepository) GetByGoogleSub(ctx context.Context, sub string) (domain.User, error) {
	query := `
		SELECT id, email, username
		FROM users
		WHERE google_sub = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, sub).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("get user by google sub: %w", err)
	}

	return user, nil
}

// LinkGoogleSub attaches a Google identity to an existing account
func (r *UserRepository) LinkGoogleSub(ctx context.Context, userID, sub string) error {
	query := `UPDATE users SET google_sub = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, sub, userID)
	if err != nil {
		return fmt.Errorf("link google sub: %w", err)
	}

	return nil
}

// CreateGoogleUser inserts a brand-new Google-only account: no password
func (r *UserRepository) CreateGoogleUser(ctx context.Context, email, username, sub string) (string, error) {
	query := `
		INSERT INTO users (email, username, google_sub)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, email, username, sub).Scan(&id)
	if err != nil {
		// 23505 = a concurrent request inserted this email first; report it as
		// ErrEmailExists instead of a raw 500.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", domain.ErrEmailExists
		}

		return "", fmt.Errorf("create google user: %w", err)
	}

	return id, nil
}
