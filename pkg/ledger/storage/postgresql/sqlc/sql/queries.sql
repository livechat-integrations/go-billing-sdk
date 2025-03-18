-- name: CreateEvent :exec
INSERT INTO events(id, lc_organization_id, type, action, payload, error, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: CreateCharge :exec
INSERT INTO charges(id, amount, status, lc_organization_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW());

-- name: UpdateChargeStatus :exec
UPDATE charges
SET status = $1, updated_at = now()
WHERE id = $2
;

-- name: CreateTopUp :exec
INSERT INTO top_ups(id, status, amount, type, lc_organization_id, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW());

-- name: UpsertTopUp :one
INSERT INTO top_ups(id, status, amount, type, lc_organization_id, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (id, lc_organization_id) DO UPDATE SET lc_charge = EXCLUDED.lc_charge, status = EXCLUDED.status, confirmation_url = EXCLUDED.confirmation_url, current_topped_up_at = EXCLUDED.current_topped_up_at, next_top_up_at = EXCLUDED.next_top_up_at, updated_at = NOW()
RETURNING *;

-- name: GetTopUpByIDAndTypeWhereStatusIsNot :one
SELECT *
FROM top_ups
WHERE id = $1
  AND type = $2
  AND status != $3
;

-- name: GetTopUpByID :one
SELECT *
FROM top_ups
WHERE id = $1
;

-- name: UpdateTopUpRequestStatus :exec
UPDATE top_ups
SET status = $1, updated_at = now()
WHERE id = $2
;

-- name: GetTopUpsByOrganizationID :many
SELECT *
FROM top_ups
WHERE lc_organization_id = $1;

-- name: GetTopUpsByOrganizationIDAndStatus :many
SELECT *
FROM top_ups
WHERE lc_organization_id = $1
  AND status = $2;

-- name: GetOrganizationBalance :one
SELECT b.amount::numeric FROM (SELECT (
    SELECT SUM(tu.amount)
    FROM top_ups tu
    WHERE tu.lc_organization_id = $1
      AND tu.status = $2
) - (
    SELECT SUM(c.amount)
    FROM charges c
    WHERE c.lc_organization_id = $1
      AND c.status = $3
) AS amount) AS b
;