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
WHERE s.lc_organization_id = $1
ORDER BY s.created_at DESC;

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
WHERE payload->>'status' = ANY($1::text[])
AND deleted_at IS NULL;

-- name: DeleteSubscription :exec
UPDATE subscriptions
SET deleted_at = NOW()
WHERE id = $1 and lc_organization_id = $2;

-- name: CreateTrialUsage :exec
INSERT INTO trial_usage (lc_organization_id)
VALUES ($1)
ON CONFLICT (lc_organization_id) DO NOTHING;

-- name: HasTrialUsage :one
SELECT EXISTS(
    SELECT 1 FROM trial_usage
    WHERE lc_organization_id = $1
);

-- name: IncrementChargeSyncErrorCount :exec
UPDATE charges
SET sync_error_count = sync_error_count + 1,
    last_sync_error_at = NOW()
WHERE id = $1
AND deleted_at IS NULL;

-- name: GetChargesWithHighErrorCount :many
SELECT *
FROM charges
WHERE sync_error_count >= $1
AND deleted_at IS NULL;