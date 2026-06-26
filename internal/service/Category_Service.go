package service

import (
	"context"

	"folosy-backend/internal/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category domain.Category) (domain.Category, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Category, error)
	GetByID(ctx context.Context, id, userID string) (domain.Category, error)
	Update(ctx context.Context, category domain.Category) (domain.Category, error)
	Delete(ctx context.Context, id, userID string) error
}

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, userID, name, icon string, budgetLimitMinor int64) (domain.Category, error) {
	category := domain.Category{
		UserID:           userID,
		Name:             name,
		Icon:             icon,
		BudgetLimitMinor: budgetLimitMinor,
	}
	return s.repo.Create(ctx, category)
}

func (s *CategoryService) List(ctx context.Context, userID string) ([]domain.Category, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *CategoryService) Get(ctx context.Context, id, userID string) (domain.Category, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *CategoryService) Update(ctx context.Context, id, userID, name, icon string, budgetLimitMinor int64) (domain.Category, error) {
	category := domain.Category{
		ID:               id,
		UserID:           userID,
		Name:             name,
		Icon:             icon,
		BudgetLimitMinor: budgetLimitMinor,
	}
	return s.repo.Update(ctx, category)
}

func (s *CategoryService) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
