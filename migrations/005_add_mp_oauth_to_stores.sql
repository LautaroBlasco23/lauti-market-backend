-- Add MercadoPago OAuth fields to stores table
ALTER TABLE stores
    ADD COLUMN IF NOT EXISTS mp_user_id VARCHAR(50),
    ADD COLUMN IF NOT EXISTS mp_access_token TEXT,
    ADD COLUMN IF NOT EXISTS mp_refresh_token TEXT,
    ADD COLUMN IF NOT EXISTS mp_token_expires_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS mp_connected_at TIMESTAMP;

-- Add index for faster lookups by mp_user_id
CREATE INDEX IF NOT EXISTS idx_stores_mp_user_id ON stores(mp_user_id);
