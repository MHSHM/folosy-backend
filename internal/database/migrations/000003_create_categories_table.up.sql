CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    budget_limit_minor BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Unique on the PAIR: one user can't have two categories with the same name,
    -- but different users can each have their own "Groceries".
    UNIQUE (user_id, name)
);

-- Our hot query is "list this user's categories" (WHERE user_id = $1), so index it explicitly.
CREATE INDEX idx_categories_user_id ON categories (user_id);