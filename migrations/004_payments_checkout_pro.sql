ALTER TABLE payments ADD COLUMN preference_id VARCHAR(255);
ALTER TABLE payments ALTER COLUMN idempotency_key DROP NOT NULL;
ALTER TABLE payments ALTER COLUMN idempotency_key SET DEFAULT NULL;

CREATE INDEX idx_payments_preference_id ON payments(preference_id);
