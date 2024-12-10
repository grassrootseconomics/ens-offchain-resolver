-- Keystore
CREATE TABLE IF NOT EXISTS alias (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    blockchain_address TEXT NOT NULL,
    primary_name TEXT NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS blockchain_address_idx ON alias(blockchain_address);
CREATE INDEX IF NOT EXISTS primary_name_idx ON alias(primary_name);