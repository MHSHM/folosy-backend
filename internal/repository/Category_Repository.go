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

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create inserts a new category for category.UserID and fills in the
// DB-generated fields (id, timestamps) on the returned copy.
func (r *CategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	query := `
		INSERT INTO categories (user_id, name, icon, budget_limit_minor)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		category.UserID, category.Name, category.Icon, category.BudgetLimitMinor,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		// 23505 here = the UNIQUE (user_id, name) constraint: this user already
		// has a category with this name.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Category{}, domain.ErrCategoryNameExists
		}
		return domain.Category{}, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

// Delete a category specified by the id and user id
func (r *CategoryRepository) Delete(ctx context.Context, id, userID string) error {
	query := `
		DELETE FROM categories WHERE id = $1 AND user_id = $2
	`
	tag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

// ListByUser returns all of a user's categories, oldest first.
func (r *CategoryRepository) ListByUser(ctx context.Context, userID string) ([]domain.Category, error) {
	query := `
		SELECT id, user_id, name, icon, budget_limit_minor, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	// rows holds an open connection; Close releases it back to the pool. Always.
	defer rows.Close()

	// Capacity hint: a SELECT streams rows, so the exact count isn't known until
	// iteration ends — but most users have only a handful of categories, so we
	// pre-size to avoid append's repeated regrow-and-copy in the common case.
	categories := make([]domain.Category, 0, 16)
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Icon, &c.BudgetLimitMinor, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	// The loop hides errors (a mid-stream network failure ends iteration without
	// panicking); rows.Err() surfaces them after the loop.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

// GetByID fetches one category, scoped to its owner. A category that doesn't
// exist OR belongs to someone else both yield ErrCategoryNotFound.
func (r *CategoryRepository) GetByID(ctx context.Context, id, userID string) (domain.Category, error) {
	query := `
		SELECT id, user_id, name, icon, budget_limit_minor, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2
	`
	var c domain.Category
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Icon, &c.BudgetLimitMinor, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, domain.ErrCategoryNotFound
		}
		return domain.Category{}, fmt.Errorf("get category by id: %w", err)
	}
	return c, nil
}

// Update changes a category's mutable fields (name, icon, budget) and bumps
// updated_at, scoped to the owner. It returns the refreshed row.
func (r *CategoryRepository) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	query := `
		UPDATE categories
		SET name = $1, icon = $2, budget_limit_minor = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, name, icon, budget_limit_minor, created_at, updated_at
	`
	var c domain.Category
	err := r.db.QueryRow(ctx, query,
		category.Name, category.Icon, category.BudgetLimitMinor, category.ID, category.UserID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Icon, &c.BudgetLimitMinor, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		// No matching row → the category doesn't exist or isn't this user's.
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, domain.ErrCategoryNotFound
		}
		// 23505 → a rename collided with another of this user's categories.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Category{}, domain.ErrCategoryNameExists
		}
		return domain.Category{}, fmt.Errorf("update category: %w", err)
	}
	return c, nil
}
