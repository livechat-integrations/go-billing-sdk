CREATE TABLE IF NOT EXISTS charges
(
    id              varchar(36) PRIMARY KEY,
    lc_organization_id varchar(36) NOT NULL,
    type            varchar(255) NOT NULL,
    payload         jsonb        NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS subscriptions
(
    id              varchar(36) PRIMARY KEY,
    lc_organization_id varchar(36) NOT NULL,
    plan_name       varchar(255) NOT NULL,
    charge_id       varchar(36) REFERENCES charges (id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE VIEW active_subscriptions AS
SELECT * FROM subscriptions WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS billing_events
(
    id                 varchar(36) NOT NULL,
    lc_organization_id varchar(36) NOT NULL,
    type               varchar(255) NOT NULL,
    action             varchar(255) NOT NULL,
    payload            jsonb,
    error              varchar(255),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
    );
ALTER TABLE billing_events
    ADD CONSTRAINT billing_events_pkey UNIQUE (id, action);
CREATE INDEX ON billing_events (lc_organization_id);