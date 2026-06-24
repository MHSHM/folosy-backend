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
