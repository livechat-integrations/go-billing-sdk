CREATE TABLE IF NOT EXISTS charges
(
    id                 varchar(36) PRIMARY KEY,
    amount             numeric(9,3) NOT NULL,
    lc_organization_id varchar(36) NOT NULL,
    status             varchar(255) NOT NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ
);
CREATE INDEX ON charges (status);
CREATE INDEX ON charges (lc_organization_id);

CREATE TABLE IF NOT EXISTS top_ups
(
    id                 varchar(36) PRIMARY KEY,
    amount             numeric(9,3) NOT NULL,
    lc_organization_id varchar(36) NOT NULL,
    type               varchar(255) NOT NULL,
    status             varchar(255) NOT NULL,
    lc_charge          jsonb,
    confirmation_url   varchar(255) NOT NULL,
    current_topped_up_at TIMESTAMPTZ,
    next_top_up_at     TIMESTAMPTZ,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ
);
CREATE INDEX ON top_ups (status);
CREATE INDEX ON top_ups (lc_organization_id);

CREATE TABLE IF NOT EXISTS events
(
    id                 varchar(36) PRIMARY KEY,
    lc_organization_id varchar(36) NOT NULL,
    type               varchar(255) NOT NULL,
    action             varchar(255) NOT NULL,
    payload            jsonb,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX ON events (lc_organization_id);