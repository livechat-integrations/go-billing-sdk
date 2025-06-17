CREATE TABLE IF NOT EXISTS ledger_ledger
(
    id                 varchar(255) UNIQUE PRIMARY KEY,
    amount             numeric(9,3) NOT NULL,
    lc_organization_id varchar(36) NOT NULL,
    payload            jsonb,
    is_voucher         boolean NOT NULL DEFAULT FALSE,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
    );
CREATE INDEX ON ledger_ledger (lc_organization_id);

CREATE TABLE IF NOT EXISTS ledger_top_ups
(
    id                 varchar(36) UNIQUE PRIMARY KEY,
    amount             numeric(9,3) NOT NULL,
    lc_organization_id varchar(36) NOT NULL,
    type               varchar(255) NOT NULL,
    status             varchar(255) NOT NULL,
    lc_charge          jsonb,
    confirmation_url   varchar(255) NOT NULL,
    current_topped_up_at TIMESTAMPTZ DEFAULT NULL,
    next_top_up_at     TIMESTAMPTZ DEFAULT NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ
    );
CREATE INDEX ON ledger_top_ups (status);
CREATE INDEX ON ledger_top_ups (lc_organization_id);

CREATE TABLE IF NOT EXISTS ledger_events
(
    id                 varchar(36) NOT NULL,
    lc_organization_id varchar(36) NOT NULL,
    type               varchar(255) NOT NULL,
    action             varchar(255) NOT NULL,
    payload            jsonb,
    error              varchar(255),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
    );
ALTER TABLE ledger_events
    ADD CONSTRAINT ledger_events_pkey UNIQUE (id, action);
CREATE INDEX ON ledger_events (lc_organization_id);