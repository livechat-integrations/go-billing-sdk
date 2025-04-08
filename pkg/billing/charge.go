package billing

import (
	"encoding/json"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

type ChargeType string

const (
	ChargeTypeRecurring ChargeType = "recurring"
)

// Charge is a representation of a charge
type Charge struct {
	ID               string
	LCOrganizationID string
	Type             ChargeType
	Payload          json.RawMessage
	CreatedAt        time.Time
	CanceledAt       *time.Time
}
type Subscription struct {
	ID               string
	Charge           *Charge
	LCOrganizationID string
	PlanName         string
	CreatedAt        time.Time
	DeletedAt        *time.Time
}

type Plan struct {
	Name   string
	Config json.RawMessage
}

type Plans []Plan

func (p Plans) GetPlan(name string) *Plan {
	for _, plan := range p {
		if plan.Name == name {
			return &plan
		}
	}
	return nil
}

func (c Subscription) IsActive() bool {
	if c.Charge == nil {
		return true
	}

	var p livechat.BaseCharge
	_ = json.Unmarshal(c.Charge.Payload, &p)
	return c.Charge.CanceledAt == nil && (p.Status == "active" || p.Status == "success")
}
