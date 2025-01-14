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

CREATE UNIQUE INDEX subscriptions_lc_organization_id_uq on subscriptions (lc_organization_id);

CREATE VIEW active_subscriptions AS
SELECT * FROM subscriptions WHERE deleted_at IS NULL;