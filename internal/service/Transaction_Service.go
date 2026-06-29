package service

import (
	"context"
	"time"

	"folosy-backend/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, t domain.Transaction) (domain.Transaction, error)
	GetByID(ctx context.Context, id, userID string) (domain.Transaction, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Transaction, error)
	TopExpenses(ctx context.Context, userID string, from, to time.Time) ([]domain.Transaction, error)
	Update(ctx context.Context, t domain.Transaction) (domain.Transaction, error)
	Delete(ctx context.Context, id, userID string) error
}

type TransactionService struct {
	repo TransactionRepository
}

func NewTransactionService(repo TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// Create is the shared write path: manual entry today, the SMS parser later.
func (s *TransactionService) Create(ctx context.Context, userID string, categoryID domain.NullString, amountMinor int64, direction domain.Direction, merchant string, occurredAt time.Time) (domain.Transaction, error) {
	t := domain.Transaction{
		UserID:      userID,
		CategoryID:  categoryID,
		AmountMinor: amountMinor,
		Direction:   direction,
		Merchant:    merchant,
		OccurredAt:  occurredAt,
	}
	return s.repo.Create(ctx, t)
}

func (s *TransactionService) List(ctx context.Context, userID string) ([]domain.Transaction, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *TransactionService) Get(ctx context.Context, id, userID string) (domain.Transaction, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *TransactionService) TopExpenses(ctx context.Context, userID string, from, to time.Time) ([]domain.Transaction, error) {
	return s.repo.TopExpenses(ctx, userID, from, to)
}

func (s *TransactionService) Update(ctx context.Context, id, userID string, categoryID domain.NullString, amountMinor int64, direction domain.Direction, merchant string, occurredAt time.Time) (domain.Transaction, error) {
	t := domain.Transaction{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		AmountMinor: amountMinor,
		Direction:   direction,
		Merchant:    merchant,
		OccurredAt:  occurredAt,
	}
	return s.repo.Update(ctx, t)
}

func (s *TransactionService) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
