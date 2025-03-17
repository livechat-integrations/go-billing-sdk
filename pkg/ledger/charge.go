package ledger

import (
	"time"
)

type ChargeStatus string

const (
	ChargeStatusActive    ChargeStatus = "active"
	ChargeStatusCancelled ChargeStatus = "cancelled"
)

type Charge struct {
	ID               string       `json:"id"`
	LCOrganizationID string       `json:"lc_organization_id"`
	Amount           float32      `json:"amount"`
	Status           ChargeStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}
