ALTER TABLE charges ADD COLUMN sync_error_count INT NOT NULL DEFAULT 0;
ALTER TABLE charges ADD COLUMN last_sync_error_at DATETIME;
CREATE INDEX idx_charges_sync_error_count ON charges(sync_error_count);
