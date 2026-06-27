CREATE TABLE transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id)      ON DELETE CASCADE,
    -- NULL = uncategorized (a valid, routine state). ON DELETE SET NULL so that
    -- deleting a category NEVER deletes its transactions
    category_id  UUID          REFERENCES categories(id) ON DELETE SET NULL,
    -- int64 minor units (cents).
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    -- direction: 1 = income, 2 = expense.
    direction    SMALLINT NOT NULL CHECK (direction IN (1, 2)),
    -- Free-text label/description: "Starbucks", "Uber", "Salary Deposit".
    merchant     TEXT NOT NULL DEFAULT '',
    -- When the money actually MOVED (user-supplied or SMS-parsed). Distinct from
    -- created_at = when WE recorded the row.
    occurred_at  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_user_occurred ON transactions (user_id, occurred_at DESC);
