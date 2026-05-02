CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked    BOOLEAN     NOT NULL DEFAULT FALSE
);

-- Index for the hot path: looking up a token string on every /auth/refresh call.
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token   ON refresh_tokens (token);
-- Index for revoking all tokens belonging to a user on logout.
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);