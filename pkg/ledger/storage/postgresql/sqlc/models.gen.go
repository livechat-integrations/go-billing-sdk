// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Charge struct {
	ID               string
	Amount           pgtype.Numeric
	LcOrganizationID string
	Status           string
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
}

type Event struct {
	ID               string
	LcOrganizationID string
	Type             string
	Action           string
	Payload          []byte
	CreatedAt        pgtype.Timestamptz
}

type TopUp struct {
	ID                string
	Amount            pgtype.Numeric
	LcOrganizationID  string
	Type              string
	Status            string
	LcCharge          []byte
	ConfirmationUrl   string
	CurrentToppedUpAt pgtype.Timestamptz
	NextTopUpAt       pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
}
