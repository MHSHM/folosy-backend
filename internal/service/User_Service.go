package service

import (
	"context"
	"fmt"
	"time"

	"folosy-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Register(ctx context.Context, user domain.User) (string, error)
	EmailExist(ctx context.Context, email string) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo}
}

func (s *UserService) Register(ctx context.Context, email, username, password string) (domain.User, error) {
	err := s.repo.EmailExist(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

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
