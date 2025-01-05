CREATE TABLE IF NOT EXISTS charges
(
    id              varchar(36) PRIMARY KEY,
    type            varchar(255) NOT NULL,
    payload         jsonb        NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS installation_charges (
    installation_id VARCHAR NOT NULL REFERENCES installations (lc_organization_id) ON DELETE CASCADE,
    charge_id       VARCHAR REFERENCES charges (id) ON DELETE CASCADE,
    PRIMARY KEY (installation_id)
);