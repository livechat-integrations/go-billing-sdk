CREATE TABLE IF NOT EXISTS charges
(
    id              VARCHAR(36),
    lc_organization_id VARCHAR(36) NOT NULL,
    type            VARCHAR(255) NOT NULL,
    payload         JSON        NOT NULL,
    created_at      DATETIME  NOT NULL DEFAULT NOW(),
    deleted_at      DATETIME,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS subscriptions
(
    id              VARCHAR(36),
    lc_organization_id VARCHAR(36) NOT NULL,
    plan_name       VARCHAR(255) NOT NULL,
    charge_id       VARCHAR(36),
    created_at      DATETIME  NOT NULL DEFAULT NOW(),
    deleted_at      DATETIME,
    PRIMARY KEY (id),
    FOREIGN KEY (charge_id)
        REFERENCES charges(id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_0900_ai_ci;

CREATE VIEW active_subscriptions AS
SELECT * FROM subscriptions WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS billing_events
(
    id                 VARCHAR(36) NOT NULL,
    lc_organization_id VARCHAR(36) NOT NULL,
    type               VARCHAR(255) NOT NULL,
    action             VARCHAR(255) NOT NULL,
    payload            JSON,
    error              VARCHAR(255),
    created_at         DATETIME  NOT NULL DEFAULT NOW(),
    UNIQUE KEY `billing_events_pkey` (`id`, `action`),
    INDEX (`lc_organization_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_0900_ai_ci;