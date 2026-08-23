-- Credential aggregate root (internal/identity/domain/credential).
CREATE TABLE IF NOT EXISTS credentials (
    id            VARCHAR(64)  PRIMARY KEY,
    -- Email VO normalizes to lowercase on construction, so a plain UNIQUE is enough.
    email         VARCHAR(255) NOT NULL UNIQUE,
    -- bcrypt hash (60 chars today); never stores plaintext.
    password_hash VARCHAR(255) NOT NULL,
    status        VARCHAR(20)  NOT NULL
        CONSTRAINT credentials_status_check CHECK (status IN ('inactive', 'active', 'suspended')),
    created_at    TIMESTAMPTZ  NOT NULL,
    updated_at    TIMESTAMPTZ  NOT NULL
);

-- List() paginates by id, but status filtering is the expected first access pattern.
CREATE INDEX IF NOT EXISTS credentials_status_idx ON credentials (status);
