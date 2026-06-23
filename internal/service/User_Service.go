package service

import (
	"folosy-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"

	"time"
)

type UserRepository interface {
	Register(user domain.User) (string, error)
	EmailExist(email string) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo}
}

func (s *UserService) Register(email, username, password string) (domain.User, error) {
	err := s.repo.EmailExist(email)
	if err != nil {
		return domain.User{}, err
	}

	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Email:    email,
		Username: username,
		Password: string(hashed_password),
	}

	id, err := s.repo.Register(user)
	if err != nil {
		return domain.User{}, err
	}

	user.ID = id
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return user, nil
}
