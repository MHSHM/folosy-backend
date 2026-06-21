package service

import (
	"folosy-backend/internal/domain"
)

type UserRepository interface {
	CreateUser(user domain.User) error
	EmailExist(email string) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo}
}

func (s *UserService) CreateUser(email, username, password string) (domain.User, error) {
	err := s.repo.EmailExist(email)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Email:    email,
		Username: username,
		Password: password,
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
