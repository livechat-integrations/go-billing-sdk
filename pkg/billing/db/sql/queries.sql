-- name: CreateCharge :exec
INSERT INTO charges(id, type, payload, created_at, deleted_at)
VALUES ($1, $2, $3, NOW(), $4);

-- name: CreateInstallationCharge :exec
INSERT INTO installation_charges(installation_id, charge_id)
VALUES ($1, $2) ON CONFLICT (installation_id) DO UPDATE SET charge_id = $2;

-- name: GetChargeByID :one
SELECT *
FROM charges c LEFT JOIN installation_charges ic on c.id = ic.charge_id
WHERE id = $1;

-- name: GetChargeByInstallationID :one
SELECT *
FROM installation_charges ic LEFT JOIN charges c on c.id = ic.charge_id
WHERE installation_id = $1;

-- name: UpdateCharge :exec
UPDATE charges
SET payload = $2
WHERE id = $1;

-- name: DeleteChargeForInstallationID :exec
UPDATE charges
SET deleted_at = now()
WHERE id = (SELECT charge_id from installation_charges WHERE installation_id = $1);
