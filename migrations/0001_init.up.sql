CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    discord_id BIGINT UNIQUE NOT NULL,
    discord_username VARCHAR(64) NOT NULL,
    hytale_username VARCHAR(64),
    hytale_uuid VARCHAR(64),
    verification_code VARCHAR(16),
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_verification_code ON users (verification_code);
CREATE INDEX IF NOT EXISTS idx_users_hytale_username ON users (hytale_username);
CREATE INDEX IF NOT EXISTS idx_users_hytale_uuid ON users (hytale_uuid);
