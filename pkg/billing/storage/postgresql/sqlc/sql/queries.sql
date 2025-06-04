-- name: CreateCharge :exec
INSERT INTO charges(id, type, payload, lc_organization_id, created_at)
VALUES ($1, $2, $3, $4, NOW());

-- name: GetChargeByID :one
SELECT *
FROM charges
WHERE id = $1
AND deleted_at IS NULL;

-- name: GetChargeByOrganizationID :one
SELECT *
FROM charges
WHERE lc_organization_id = $1
AND deleted_at IS NULL;

-- name: UpdateCharge :exec
UPDATE charges
SET payload = $2
WHERE id = $1
AND deleted_at IS NULL;


-- name: CreateSubscription :exec
INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at)
VALUES ($1, $2, $3, $4, NOW());

-- name: GetSubscriptionsByOrganizationID :many
SELECT *
FROM active_subscriptions s
LEFT JOIN charges c on s.charge_id = c.id
WHERE s.lc_organization_id = $1;

-- name: GetSubscriptionByChargeID :one
SELECT *
FROM active_subscriptions
WHERE charge_id = $1;

-- name: DeleteSubscriptionByChargeID :exec
UPDATE subscriptions
SET deleted_at = now()
WHERE charge_id = $1
AND lc_organization_id = $2;

-- name: DeleteCharge :exec
UPDATE charges
SET deleted_at = now()
WHERE id = $1;

-- name: GetChargesByOrganizationID :many
SELECT *
FROM charges
WHERE lc_organization_id = $1;

-- name: CreateEvent :exec
INSERT INTO billing_events(id, lc_organization_id, type, action, payload, error, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: GetChargesByStatuses :many
SELECT *
FROM charges
WHERE (payload->>'status') = ANY($1::text[]);