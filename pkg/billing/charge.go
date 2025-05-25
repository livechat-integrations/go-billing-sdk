package billing

import (
	"encoding/json"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/internal/livechat"
)

type ChargeType string

const (
	ChargeTypeRecurring ChargeType = "recurring"

	RetentionPeriod = 4 * 24 * time.Hour
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

	var p livechat.RecurrentCharge
	_ = json.Unmarshal(c.Charge.Payload, &p)
	if p.NextChargeAt == nil {
		return false
	}

	return c.Charge.CanceledAt == nil &&
		p.NextChargeAt.Add(RetentionPeriod).After(time.Now()) &&
		(p.Status == "active" || p.Status == "past_due")
}

func GetSyncValidStatuses() []string {
	return []string{"active", "pending", "accepted"}
}
