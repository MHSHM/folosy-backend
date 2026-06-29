CREATE TABLE budgets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- int64 minor units (cents), same money model as everywhere else.
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    -- The budget's window. Spent = SUM of expenses dated within [starts_at, ends_at).
    starts_at    TIMESTAMPTZ NOT NULL,
    ends_at      TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- A budget can't end before it begins.
    CHECK (ends_at > starts_at)
);

CREATE INDEX idx_budgets_user_starts ON budgets (user_id, starts_at DESC);