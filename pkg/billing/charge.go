package billing

import (
	"encoding/json"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
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
	Status           livechat.ChargeStatus
	Payload          json.RawMessage
	CurrentChargeAt  *time.Time
	NextChargeAt     *time.Time
	CreatedAt        time.Time
	CanceledAt       *time.Time
	SyncErrorCount   int
	LastSyncErrorAt  *time.Time
}
type Subscription struct {
	ID               string
	Charge           *Charge
	LCOrganizationID string
	PlanName         string
	DunningEndDate   *time.Time
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

	// During trial period: has TrialEndsAt and no CurrentChargeAt yet
	if p.TrialEndsAt != nil && p.CurrentChargeAt == nil {
		// Active if still in trial period
		if p.TrialEndsAt.After(time.Now()) {
			return true
		}
		return false
	}

	// For regular subscriptions: NextChargeAt is set but CurrentChargeAt is nil means pending/not active
	if p.NextChargeAt != nil && p.CurrentChargeAt == nil {
		return false
	}

	return c.Charge.CanceledAt == nil &&
		p.NextChargeAt.Add(RetentionPeriod).After(time.Now()) &&
		(p.Status == "active" || p.Status == "past_due")
}

func (c Subscription) IsTrialActive() bool {
	if c.Charge == nil {
		return false
	}

	var p livechat.RecurrentCharge
	_ = json.Unmarshal(c.Charge.Payload, &p)

	return p.TrialEndsAt != nil &&
		p.TrialEndsAt.After(time.Now()) &&
		p.CurrentChargeAt == nil
}

func GetSyncValidStatuses() []string {
	return []string{"active", "pending", "accepted", "past_due", "frozen"}
}
