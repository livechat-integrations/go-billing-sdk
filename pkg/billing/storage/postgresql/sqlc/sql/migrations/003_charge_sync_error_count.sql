ALTER TABLE charges ADD COLUMN sync_error_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE charges ADD COLUMN last_sync_error_at TIMESTAMPTZ;
CREATE INDEX idx_charges_sync_error_count ON charges(sync_error_count) WHERE sync_error_count >= 10 AND deleted_at IS NULL;
