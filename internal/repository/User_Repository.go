package repository

import (
	"context"
	"fmt"

	"folosy-backend/internal/domain"

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
		return "", fmt.Errorf("register new user: %w", err)
	}

	return id, nil
}

func (r *UserRepository) EmailExist(ctx context.Context, email string) error {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check email exists: %w", err)
	}

	if exists {
		return domain.ErrEmailExists
	}

	return nil
}
