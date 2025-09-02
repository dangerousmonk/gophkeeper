CREATE TYPE vault_type AS ENUM ('credentials', 'text', 'binary', 'bank_card');
CREATE TABLE IF NOT EXISTS vault (
    id  BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR NOT NULL,
    data_type vault_type NOT NULL,
    encrypted_data BYTEA NOT NULL,
    meta_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_vault_user_id ON vault(user_id);