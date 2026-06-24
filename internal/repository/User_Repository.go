package repository

import (
	"context"
	"errors"
	"fmt"

	"folosy-backend/internal/domain"

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
