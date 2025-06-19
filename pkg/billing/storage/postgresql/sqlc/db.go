//go:generate go run -mod=readonly github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

package sqlc

import (
	"encoding/json"
	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/billing"
	"time"
)

func (c *Charge) ToBillingCharge() *billing.Charge {
	var canceledAt *time.Time
	if c.DeletedAt.Valid {
		canceledAt = &c.DeletedAt.Time
	}

	var nextChargeAt *time.Time
	if c.Type == string(billing.ChargeTypeRecurring) {
		var p livechat.RecurrentCharge
		_ = json.Unmarshal(c.Payload, &p)

		nextChargeAt = p.NextChargeAt
	}

	var b livechat.BaseCharge
	_ = json.Unmarshal(c.Payload, &b)

	return &billing.Charge{
		ID:               c.ID,
		LCOrganizationID: c.LcOrganizationID,
		Type:             billing.ChargeType(c.Type),
		Status:           b.Status,
		Payload:          c.Payload,
		NextChargeAt:     nextChargeAt,
		CreatedAt:        c.CreatedAt.Time,
		CanceledAt:       canceledAt,
	}
}

func (r *GetSubscriptionsByOrganizationIDRow) ToBillingSubscription() *billing.Subscription {
	var deletedAt *time.Time
	if r.DeletedAt.Valid {
		deletedAt = &r.DeletedAt.Time
	}

	subscription := &billing.Subscription{
		ID:               r.ID,
		LCOrganizationID: r.LcOrganizationID,
		PlanName:         r.PlanName,
		CreatedAt:        r.CreatedAt.Time,
		DeletedAt:        deletedAt,
	}

	if !r.ChargeID.Valid {
		return subscription
	}

	var chargeDeletedAt *time.Time
	if r.DeletedAt_2.Valid {
		chargeDeletedAt = &r.DeletedAt_2.Time
	}

	var p livechat.RecurrentCharge
	_ = json.Unmarshal(r.Payload, &p)

	subscription.Charge = &billing.Charge{
		ID:               r.ChargeID.String,
		LCOrganizationID: r.LcOrganizationID_2.String,
		Type:             billing.ChargeType(r.Type.String),
		Payload:          r.Payload,
		NextChargeAt:     p.NextChargeAt,
		CreatedAt:        r.CreatedAt_2.Time,
		CanceledAt:       chargeDeletedAt,
		Status:           p.Status,
	}

	var dunningEndDate time.Time
	if p.Status == livechat.RecurrentChargeStatusPastDue || p.Status == livechat.RecurrentChargeStatusFrozen {
		if p.NextChargeAt != nil {
			dunningEndDate = p.NextChargeAt.AddDate(0, 0, 16)
			for dunningEndDate.Before(time.Now()) {
				dunningEndDate = dunningEndDate.AddDate(0, 0, 16)
			}
		}

		subscription.DunningEndDate = &dunningEndDate
	}

	if (p.Status == livechat.RecurrentChargeStatusActive || p.Status == livechat.RecurrentChargeStatusFrozen) && p.NextChargeAt == nil {
		dunningEndDate = p.CreatedAt.AddDate(0, 0, 16)
		for dunningEndDate.Before(time.Now()) {
			dunningEndDate = dunningEndDate.AddDate(0, 0, 16)
		}
		subscription.DunningEndDate = &dunningEndDate
	}

	return subscription
}
