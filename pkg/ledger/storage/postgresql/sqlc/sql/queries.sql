-- name: CreateEvent :exec
INSERT INTO ledger_events(id, lc_organization_id, type, action, payload, error, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: CreateLedgerOperation :exec
INSERT INTO ledger_ledger(id, amount, lc_organization_id, payload, is_voucher, created_at)
VALUES ($1, $2, $3, $4, $5, NOW());


-- name: UpsertTopUp :one
INSERT INTO ledger_top_ups(id, status, amount, type, lc_organization_id, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(),NOW())
ON CONFLICT ON CONSTRAINT ledger_top_ups_pkey DO UPDATE SET lc_charge = EXCLUDED.lc_charge, status = EXCLUDED.status, current_topped_up_at = EXCLUDED.current_topped_up_at, next_top_up_at = EXCLUDED.next_top_up_at, updated_at = NOW()
RETURNING *;

-- name: GetTopUpByIDAndTypeWhereStatusIsNot :one
SELECT *
FROM ledger_top_ups
WHERE id = $1
  AND type = $2
  AND status != $3
ORDER BY created_at DESC;

-- name: GetTopUpsByTypeWhereStatusNotIn :many
SELECT *
FROM ledger_top_ups
WHERE type = $1
  AND NOT (status = ANY($2::text[]))
ORDER BY created_at ASC
    LIMIT 200;

-- name: GetRecurrentTopUpsWhereStatusNotIn :many
SELECT *
FROM ledger_top_ups
WHERE type = 'recurrent'
  AND NOT (status = ANY($1::text[]))
  AND ((next_top_up_at IS NOT NULL AND next_top_up_at <= NOW() AND status = 'active') || status != 'active')
ORDER BY created_at ASC
    LIMIT 200;

-- name: GetLedgerOperationsByOrganizationID :many
SELECT *
FROM ledger_ledger
WHERE lc_organization_id = $1
  AND is_voucher = $2
ORDER BY created_at DESC
;

-- name: GetLedgerOperation :one
SELECT *
FROM ledger_ledger
WHERE lc_organization_id = $1
  AND id = $2
ORDER BY created_at DESC
;

-- name: GetDirectTopUpsWithoutOperations :many
SELECT tups.*
FROM ledger_top_ups tups
LEFT JOIN ledger_ledger lgr ON tups.id = lgr.id AND tups.lc_organization_id = lgr.lc_organization_id
WHERE tups.type = 'direct'
    AND tups.status = 'success'
    AND lgr.id IS NULL
LIMIT 100
;

-- name: UpdateTopUpRequestStatus :exec
UPDATE ledger_top_ups
SET status = $1, updated_at = now()
WHERE id = $2
;

-- name: GetTopUpsByOrganizationID :many
SELECT *
FROM ledger_top_ups
WHERE lc_organization_id = $1;

-- name: GetTopUpByIDAndOrganizationID :one
SELECT *
FROM ledger_top_ups
WHERE id = $1
    AND lc_organization_id = $2;

-- name: GetTopUpsByOrganizationIDAndStatus :many
SELECT *
FROM ledger_top_ups
WHERE lc_organization_id = $1
  AND status = $2;

-- name: GetOrganizationBalance :one
SELECT COALESCE(SUM(amount), 0)::numeric AS amount FROM ledger_ledger WHERE lc_organization_id = $1
;