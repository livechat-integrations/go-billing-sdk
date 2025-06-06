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

	return &billing.Charge{
		ID:               c.ID,
		LCOrganizationID: c.LcOrganizationID,
		Type:             billing.ChargeType(c.Type),
		Status:           livechat.ChargeStatus(c.Status),
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
		Status:           livechat.ChargeStatus(r.Status.String),
	}

	if p.Status == livechat.RecurrentChargeStatusPastDue {
		var dunningEndDate time.Time
		if p.NextChargeAt == nil {
			dunningEndDate = p.CreatedAt.AddDate(0, 0, 16)
		} else {
			dunningEndDate = p.NextChargeAt.AddDate(0, 0, 16)
			for dunningEndDate.After(time.Now()) {
				dunningEndDate = dunningEndDate.AddDate(0, 0, 16)
			}
		}

		subscription.DunningEndDate = &dunningEndDate
	}

	return subscription
}
