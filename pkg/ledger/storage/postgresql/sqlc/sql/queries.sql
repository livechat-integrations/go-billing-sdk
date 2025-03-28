-- name: CreateEvent :exec
INSERT INTO ledger_events(id, lc_organization_id, type, action, payload, error, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: CreateCharge :exec
INSERT INTO ledger_charges(id, amount, status, lc_organization_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW());

-- name: UpdateChargeStatus :exec
UPDATE ledger_charges
SET status = $1, updated_at = now()
WHERE id = $2
;

-- name: UpsertTopUp :one
INSERT INTO ledger_top_ups(id, status, amount, type, lc_organization_id, lc_charge, confirmation_url, current_topped_up_at, unique_at, next_top_up_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, '1970-01-01 00:00:00+00'::timestamptz), $10, COALESCE((SELECT lups.created_at FROM ledger_top_ups lups WHERE lups.id = $1::varchar ORDER BY lups.created_at ASC LIMIT 1)::timestamptz, NOW()),NOW())
ON CONFLICT ON CONSTRAINT ledger_top_ups_pkey DO UPDATE SET lc_charge = EXCLUDED.lc_charge, status = EXCLUDED.status, updated_at = NOW()
RETURNING *;

-- name: GetTopUpByIDAndTypeWhereStatusIsNot :one
SELECT *
FROM ledger_top_ups
WHERE id = $1
  AND type = $2
  AND status != $3
ORDER BY unique_at DESC
;

-- name: InitTopUpRequiredValues :exec
UPDATE ledger_top_ups l
SET current_topped_up_at = $1, next_top_up_at = $2, unique_at = $3
WHERE l.id = $4
  AND l.type = $5
  AND l.status = $6
  AND l.current_topped_up_at IS NULL
;

-- name: UpdateTopUpRequestStatus :exec
UPDATE ledger_top_ups
SET status = $1, updated_at = now()
WHERE id = $2
  AND unique_at = COALESCE($3, '1970-01-01 00:00:00+00'::timestamptz)
;

-- name: GetTopUpsByOrganizationID :many
SELECT *
FROM ledger_top_ups
WHERE lc_organization_id = $1;

-- name: GetTopUpsByOrganizationIDAndStatus :many
SELECT *
FROM ledger_top_ups
WHERE lc_organization_id = $1
  AND status = $2;

-- name: GetOrganizationBalance :one
SELECT b.amount::numeric FROM (SELECT (
    SELECT COALESCE(SUM(tu.amount), 0)
    FROM ledger_top_ups tu
    WHERE tu.lc_organization_id = $1
      AND tu.status = $2
) - (
    SELECT COALESCE(SUM(c.amount), 0)
    FROM ledger_charges c
    WHERE c.lc_organization_id = $1
      AND c.status = $3
) AS amount) AS b
;