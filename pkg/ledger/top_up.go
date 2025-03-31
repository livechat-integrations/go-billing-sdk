package ledger

import (
	"encoding/json"
	"time"
)

type TopUpType string

const (
	TopUpTypeDirect    TopUpType = "direct"
	TopUpTypeRecurrent TopUpType = "recurrent"
)

type TopUpStatus string

const (
	TopUpStatusPending   TopUpStatus = "pending"
	TopUpStatusActive    TopUpStatus = "active"
	TopUpStatusCancelled TopUpStatus = "cancelled"
	TopUpStatusFailed    TopUpStatus = "failed"
	TopUpStatusDeclined  TopUpStatus = "declined"
)

type TopUp struct {
	ID                string          `json:"id"`
	LCOrganizationID  string          `json:"lc_organization_id"`
	Status            TopUpStatus     `json:"status"`
	Amount            float32         `json:"amount"`
	Type              TopUpType       `json:"type"`
	ConfirmationUrl   string          `json:"confirmation_url"`
	CurrentToppedUpAt *time.Time      `json:"current_topped_up_at"`
	NextTopUpAt       *time.Time      `json:"next_top_up_at"`
	LCCharge          json.RawMessage `json:"lc_charge"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
