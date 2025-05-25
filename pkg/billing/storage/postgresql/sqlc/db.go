//go:generate go run -mod=readonly github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

package sqlc

import (
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing"
	"time"
)

func (c *Charge) ToBillingCharge() *billing.Charge {
	var canceledAt *time.Time
	if c.DeletedAt.Valid {
		canceledAt = &c.DeletedAt.Time
	}

	return &billing.Charge{
		ID:               c.ID,
		LCOrganizationID: c.LcOrganizationID,
		Type:             billing.ChargeType(c.Type),
		Payload:          c.Payload,
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

	subscription.Charge = &billing.Charge{
		ID:               r.ChargeID.String,
		LCOrganizationID: r.LcOrganizationID_2.String,
		Type:             billing.ChargeType(r.Type.String),
		Payload:          r.Payload,
		CreatedAt:        r.CreatedAt_2.Time,
		CanceledAt:       chargeDeletedAt,
	}

	return subscription
}
