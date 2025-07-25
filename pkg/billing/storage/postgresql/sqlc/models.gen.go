// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type ActiveSubscription struct {
	ID               string
	LcOrganizationID string
	PlanName         string
	ChargeID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

type BillingEvent struct {
	ID               string
	LcOrganizationID string
	Type             string
	Action           string
	Payload          []byte
	Error            pgtype.Text
	CreatedAt        pgtype.Timestamptz
}

type Charge struct {
	ID               string
	LcOrganizationID string
	Type             string
	Payload          []byte
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

type Subscription struct {
	ID               string
	LcOrganizationID string
	PlanName         string
	ChargeID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}
