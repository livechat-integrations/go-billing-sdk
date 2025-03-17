// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createCharge = `-- name: CreateCharge :exec
INSERT INTO charges(id, amount, status, lc_organization_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
`

type CreateChargeParams struct {
	ID               string
	Amount           pgtype.Numeric
	Status           string
	LcOrganizationID string
}

func (q *Queries) CreateCharge(ctx context.Context, arg CreateChargeParams) error {
	_, err := q.db.Exec(ctx, createCharge,
		arg.ID,
		arg.Amount,
		arg.Status,
		arg.LcOrganizationID,
	)
	return err
}

const createEvent = `-- name: CreateEvent :exec
INSERT INTO events(id, lc_organization_id, type, action, payload, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
`

type CreateEventParams struct {
	ID               string
	LcOrganizationID string
	Type             string
	Action           string
	Payload          []byte
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) error {
	_, err := q.db.Exec(ctx, createEvent,
		arg.ID,
		arg.LcOrganizationID,
		arg.Type,
		arg.Action,
		arg.Payload,
	)
	return err
}

const createTopUp = `-- name: CreateTopUp :exec
INSERT INTO top_ups(id, status, amount, type, lc_organization_id, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
`

type CreateTopUpParams struct {
	ID                string
	Status            string
	Amount            pgtype.Numeric
	Type              string
	LcOrganizationID  string
	LcCharge          []byte
	ConfirmationUrl   string
	CurrentToppedUpAt pgtype.Timestamptz
	NextTopUpAt       pgtype.Timestamptz
}

func (q *Queries) CreateTopUp(ctx context.Context, arg CreateTopUpParams) error {
	_, err := q.db.Exec(ctx, createTopUp,
		arg.ID,
		arg.Status,
		arg.Amount,
		arg.Type,
		arg.LcOrganizationID,
		arg.LcCharge,
		arg.ConfirmationUrl,
		arg.CurrentToppedUpAt,
		arg.NextTopUpAt,
	)
	return err
}

const getChargeByIDWhereStatusIsNot = `-- name: GetChargeByIDWhereStatusIsNot :one
SELECT id, amount, lc_organization_id, status, created_at, updated_at
FROM charges
WHERE id = $1
  AND status != $2
`

type GetChargeByIDWhereStatusIsNotParams struct {
	ID     string
	Status string
}

func (q *Queries) GetChargeByIDWhereStatusIsNot(ctx context.Context, arg GetChargeByIDWhereStatusIsNotParams) (Charge, error) {
	row := q.db.QueryRow(ctx, getChargeByIDWhereStatusIsNot, arg.ID, arg.Status)
	var i Charge
	err := row.Scan(
		&i.ID,
		&i.Amount,
		&i.LcOrganizationID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrganizationBalance = `-- name: GetOrganizationBalance :one
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
`

type GetOrganizationBalanceParams struct {
	LcOrganizationID string
	Status           string
	Status_2         string
}

func (q *Queries) GetOrganizationBalance(ctx context.Context, arg GetOrganizationBalanceParams) (pgtype.Numeric, error) {
	row := q.db.QueryRow(ctx, getOrganizationBalance, arg.LcOrganizationID, arg.Status, arg.Status_2)
	var b_amount pgtype.Numeric
	err := row.Scan(&b_amount)
	return b_amount, err
}

const getTopUpByIDAndTypeWhereStatusIsNot = `-- name: GetTopUpByIDAndTypeWhereStatusIsNot :one
SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at
FROM top_ups
WHERE id = $1
  AND type = $2
  AND status != $3
`

type GetTopUpByIDAndTypeWhereStatusIsNotParams struct {
	ID     string
	Type   string
	Status string
}

func (q *Queries) GetTopUpByIDAndTypeWhereStatusIsNot(ctx context.Context, arg GetTopUpByIDAndTypeWhereStatusIsNotParams) (TopUp, error) {
	row := q.db.QueryRow(ctx, getTopUpByIDAndTypeWhereStatusIsNot, arg.ID, arg.Type, arg.Status)
	var i TopUp
	err := row.Scan(
		&i.ID,
		&i.Amount,
		&i.LcOrganizationID,
		&i.Type,
		&i.Status,
		&i.LcCharge,
		&i.ConfirmationUrl,
		&i.CurrentToppedUpAt,
		&i.NextTopUpAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getTopUpsByOrganizationID = `-- name: GetTopUpsByOrganizationID :many
SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at
FROM top_ups
WHERE lc_organization_id = $1
`

func (q *Queries) GetTopUpsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]TopUp, error) {
	rows, err := q.db.Query(ctx, getTopUpsByOrganizationID, lcOrganizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TopUp
	for rows.Next() {
		var i TopUp
		if err := rows.Scan(
			&i.ID,
			&i.Amount,
			&i.LcOrganizationID,
			&i.Type,
			&i.Status,
			&i.LcCharge,
			&i.ConfirmationUrl,
			&i.CurrentToppedUpAt,
			&i.NextTopUpAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateChargeStatus = `-- name: UpdateChargeStatus :exec
UPDATE charges
SET status = $1, updated_at = now()
WHERE id = $2
`

type UpdateChargeStatusParams struct {
	Status string
	ID     string
}

func (q *Queries) UpdateChargeStatus(ctx context.Context, arg UpdateChargeStatusParams) error {
	_, err := q.db.Exec(ctx, updateChargeStatus, arg.Status, arg.ID)
	return err
}

const updateTopUpRequestStatus = `-- name: UpdateTopUpRequestStatus :exec
UPDATE top_ups
SET status = $1, updated_at = now()
WHERE id = $2
`

type UpdateTopUpRequestStatusParams struct {
	Status string
	ID     string
}

func (q *Queries) UpdateTopUpRequestStatus(ctx context.Context, arg UpdateTopUpRequestStatusParams) error {
	_, err := q.db.Exec(ctx, updateTopUpRequestStatus, arg.Status, arg.ID)
	return err
}
