CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    -- NULL = active. A non-null timestamp means the token was rotated/revoked
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Reuse detection revokes ALL of a user's tokens (WHERE user_id = $1); index it.
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
