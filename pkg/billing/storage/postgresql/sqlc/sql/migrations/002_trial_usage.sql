CREATE TABLE IF NOT EXISTS trial_usage
(
    lc_organization_id varchar(36) PRIMARY KEY,
    used_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_trial_usage_used_at ON trial_usage(used_at);