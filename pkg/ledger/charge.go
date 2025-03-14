package ledger

import (
	"encoding/json"
	"time"
)

type ChargeType string

const (
	ChargeTypeDirect    ChargeType = "direct"
	ChargeTypeRecurrent ChargeType = "recurrent"
)

type ChargeStatus string

const (
	ChargeStatusPending   ChargeStatus = "pending"
	ChargeStatusActive    ChargeStatus = "active"
	ChargeStatusCancelled ChargeStatus = "cancelled"
)

type Charge struct {
	ID               string          `json:"id"`
	LCOrganizationID string          `json:"lc_organization_id"`
	Amount           float32         `json:"amount"`
	Type             ChargeType      `json:"type"`
	Status           ChargeStatus    `json:"status"`
	LCCharge         json.RawMessage `json:"lc_charge"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
