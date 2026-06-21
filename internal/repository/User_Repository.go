package repository

import (
	"context"
	"errors"
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

func (r *UserRepository) CreateUser(user domain.User) error {
	const query = `
		INSERT INTO users (email, username, password)
		VALUES ($1, $2, $3)
	`
	// TODO: We probably should pass down the context from the handler
	_, err := r.db.Exec(context.Background(), query, user.Email, user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *UserRepository) EmailExist(email string) error {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(context.Background(), query, email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check email exists: %w", err)
	}

	if exists {
		return errors.New("email already exists")
	}

	return nil
}
