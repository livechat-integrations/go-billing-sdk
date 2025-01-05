package billing

import (
	"encoding/json"
	"time"
)

type ChargeType string

const (
	ChargeTypeRecurring ChargeType = "recurring"
)

// Charge is a representation of a charge
type Charge struct {
	ID         string
	Type       ChargeType
	Payload    json.RawMessage
	CreatedAt  time.Time
	CanceledAt *time.Time
}

type InstallationCharge struct {
	InstallationID string
	Charge         *Charge
}

func (c InstallationCharge) IsActive() bool {
	var p BaseCharge
	_ = json.Unmarshal(c.Charge.Payload, &p)
	return c.Charge == nil ||
		(c.Charge.CanceledAt == nil && (p.Status == "active" || p.Status == "success"))
}
