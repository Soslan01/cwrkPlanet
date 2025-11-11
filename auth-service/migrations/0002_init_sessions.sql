CREATE TABLE IF NOT EXISTS auth_sessions (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash       TEXT         NOT NULL,
    expires_at       TIMESTAMPTZ  NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    user_agent       TEXT,
    ip               INET,

    CONSTRAINT token_hash_not_empty CHECK (length(trim(token_hash)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON auth_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON auth_sessions (expires_at);
