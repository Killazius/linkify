CREATE TABLE auth_schema.refresh_tokens (
                                token_hash TEXT PRIMARY KEY,
                                user_id BIGINT NOT NULL,
                                expires_at TIMESTAMP NOT NULL,
                                created_at TIMESTAMP DEFAULT NOW(),
                                FOREIGN KEY (user_id) REFERENCES auth_schema.users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON auth_schema.refresh_tokens(user_id);