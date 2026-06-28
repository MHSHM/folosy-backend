package repository

import (
	"context"
	"errors"
	"fmt"

	"folosy-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a transaction AND moves the user's balance in one atomic unit:
func (r *TransactionRepository) Create(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	// Translate magnitude + direction into a signed delta for the balance.
	balanceDelta := t.AmountMinor
	if t.Direction == domain.DirectionExpense {
		balanceDelta = -balanceDelta
	}

	// Begin opens a transaction. Everything we run on `tx` (not r.db) stays
	// invisible until Commit.
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// We use INSERT ... SELECT ... WHERE (not VALUES) so the insert ALSO enforces
	// category ownership in the same statement — one round trip, no race window.
	//
	//   SELECT $1..$6        builds exactly one candidate row from our params
	//                        (a SELECT with no FROM just emits its literal values).
	//   WHERE ...            is a gate on that single row: if true the row is
	//                        inserted; if false the SELECT yields 0 rows and
	//                        NOTHING is inserted (unlike VALUES, which is
	//                        unconditional).
	//
	// The gate passes when EITHER:
	//   $2 IS NULL                    the transaction is uncategorized — nothing
	//                                 to own, so always allow it through.
	//   EXISTS (... id=$2 AND         a category was given: allow ONLY if a
	//           user_id=$1)           category with that id is owned by THIS user.
	//                                 EXISTS is a yes/no presence check (SELECT 1
	//                                 is a throwaway — we want existence, not data).
	//
	// The FK alone can't do this: it proves the category exists SOMEWHERE, not
	// that it's ours. A cross-user category_id satisfies the FK but fails EXISTS.
	//
	// Consequence: "category given but not owned (or non-existent)" → 0 rows →
	// RETURNING has nothing → pgx.ErrNoRows, which we map to ErrCategoryNotFound.
	insertQuery := `
		INSERT INTO transactions (user_id, category_id, amount_minor, direction, merchant, occurred_at)
		SELECT $1, $2, $3, $4, $5, $6
		WHERE $2 IS NULL
		   OR EXISTS (SELECT 1 FROM categories WHERE id = $2 AND user_id = $1)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, insertQuery,
		t.UserID, t.CategoryID, t.AmountMinor, int16(t.Direction), t.Merchant, t.OccurredAt,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		// No row came back = the ownership gate shut: the category either doesn't
		// exist or isn't this user's. Both collapse to the same not-found error so
		// we never reveal that another user's category exists.
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, domain.ErrCategoryNotFound
		}
		return domain.Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	userUpdateQuery := `
		UPDATE users SET total_balance_minor = total_balance_minor + $1 WHERE id = $2
	`

	_, err = tx.Exec(ctx, userUpdateQuery, balanceDelta, t.UserID)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("user balance update %w", err)
	}

	// Commit both writes atomically
	if err := tx.Commit(ctx); err != nil {
		return domain.Transaction{}, fmt.Errorf("commit tx: %w", err)
	}
	return t, nil
}

// GetByID fetches one transaction, scoped to its owner. A transaction that
// doesn't exist OR belongs to someone else both yield ErrTransactionNotFound
func (r *TransactionRepository) GetByID(ctx context.Context, id, userID string) (domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount_minor, direction, merchant, occurred_at, created_at, updated_at
		FROM transactions
		WHERE id = $1 AND user_id = $2
	`
	var t domain.Transaction
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&t.ID, &t.UserID, &t.CategoryID, &t.AmountMinor, &t.Direction,
		&t.Merchant, &t.OccurredAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, domain.ErrTransactionNotFound
		}
		return domain.Transaction{}, fmt.Errorf("get transaction by id: %w", err)
	}
	return t, nil
}

// ListByUser returns all of a user's transactions, newest first by occurred_at.
func (r *TransactionRepository) ListByUser(ctx context.Context, userID string) ([]domain.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, amount_minor, direction, merchant, occurred_at, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY occurred_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]domain.Transaction, 0, 64)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.CategoryID, &t.AmountMinor, &t.Direction,
			&t.Merchant, &t.OccurredAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}
	return transactions, nil
}

// Delete removes a transaction AND reverses its baked-in balance effect, atomically.
func (r *TransactionRepository) Delete(ctx context.Context, id, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	const deleteQuery = `
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2
		RETURNING amount_minor, direction
	`
	var oldAmount int64
	var oldDirection domain.Direction
	if err := tx.QueryRow(ctx, deleteQuery, id, userID).Scan(&oldAmount, &oldDirection); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTransactionNotFound
		}
		return fmt.Errorf("delete transaction: %w", err)
	}

	reverseDelta := oldAmount
	if oldDirection == domain.DirectionIncome {
		reverseDelta = -reverseDelta
	}

	const updateBalance = `UPDATE users SET total_balance_minor = total_balance_minor + $1 WHERE id = $2`
	if _, err := tx.Exec(ctx, updateBalance, reverseDelta, userID); err != nil {
		return fmt.Errorf("reverse balance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// Update changes a transaction's fields AND adjusts the balance by the net
// difference (newEffect - oldEffect), atomically.
func (r *TransactionRepository) Update(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// We lock the row for the extreme case where another writer modifies the same row
	// at the same time which practically should never happen (a user can't double-modify the same
	// transaction at the same time)
	const selectOld = `
		SELECT amount_minor, direction
		FROM transactions
		WHERE id = $1 AND user_id = $2
		FOR UPDATE
	`
	var oldAmount int64
	var oldDirection domain.Direction
	if err := tx.QueryRow(ctx, selectOld, t.ID, t.UserID).Scan(&oldAmount, &oldDirection); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, domain.ErrTransactionNotFound
		}
		return domain.Transaction{}, fmt.Errorf("lock transaction for update: %w", err)
	}

	// step 2: write the new fields, re-checking category ownership. -- TODO
	// step 3: move the balance by newEffect - oldEffect.            -- TODO
	_ = oldAmount
	_ = oldDirection

	if err := tx.Commit(ctx); err != nil {
		return domain.Transaction{}, fmt.Errorf("commit tx: %w", err)
	}
	return t, nil
}
