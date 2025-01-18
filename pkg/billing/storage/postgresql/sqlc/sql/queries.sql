-- name: CreateCharge :exec
INSERT INTO charges(id, type, payload, lc_organization_id, created_at)
VALUES ($1, $2, $3, $4, NOW());

-- name: GetChargeByID :one
SELECT *
FROM charges
WHERE id = $1;

-- name: GetChargeByOrganizationID :one
SELECT *
FROM charges
WHERE lc_organization_id = $1;

-- name: UpdateCharge :exec
UPDATE charges
SET payload = $2
WHERE id = $1;


-- name: CreateSubscription :exec
INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at)
VALUES ($1, $2, $3, $4, NOW());

-- name: GetSubscriptionByOrganizationID :one
SELECT *
FROM subscriptions s
LEFT JOIN charges c on s.charge_id = c.id
WHERE s.lc_organization_id = $1;

-- name: DeleteSubscriptionByOrganizationID :exec
UPDATE subscriptions
SET deleted_at = now()
WHERE lc_organization_id = $1;

-- name: DeleteCharge :exec
UPDATE charges
SET deleted_at = now()
WHERE id = $1;