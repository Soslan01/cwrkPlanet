CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id             BIGSERIAL PRIMARY KEY,
    email          CITEXT UNIQUE NOT NULL,
    email_verified BOOLEAN      NOT NULL DEFAULT FALSE,
    password_hash  TEXT         NOT NULL,
    display_name   TEXT,
    avatar_url     TEXT,

    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT email_not_empty CHECK (length(trim(email::text)) > 0),
    CONSTRAINT password_hash_not_empty CHECK (length(trim(password_hash)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at);