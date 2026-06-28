CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) NOT NULL,
    password TEXT,
    google_sub TEXT UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    total_balance_minor BIGINT NOT NULL DEFAULT 0,
    budget NUMERIC(15, 2) DEFAULT 0.00,
    budget_start_at TIMESTAMP WITH TIME ZONE,
    budget_end_at TIMESTAMP WITH TIME ZONE,
    spent NUMERIC(15, 2) DEFAULT 0.00,
    bank_id VARCHAR(255)
);