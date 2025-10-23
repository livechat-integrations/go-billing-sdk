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

	var lastSyncErrorAt *time.Time
	if c.LastSyncErrorAt.Valid {
		lastSyncErrorAt = &c.LastSyncErrorAt.Time
	}

	var nextChargeAt, currentChargeAt *time.Time
	if c.Type == string(billing.ChargeTypeRecurring) {
		var p livechat.RecurrentCharge
		_ = json.Unmarshal(c.Payload, &p)

		nextChargeAt = p.NextChargeAt
		currentChargeAt = p.CurrentChargeAt
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
		CurrentChargeAt:  currentChargeAt,
		CreatedAt:        c.CreatedAt.Time,
		CanceledAt:       canceledAt,
		SyncErrorCount:   int(c.SyncErrorCount),
		LastSyncErrorAt:  lastSyncErrorAt,
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

	var lastSyncErrorAt *time.Time
	if r.LastSyncErrorAt.Valid {
		lastSyncErrorAt = &r.LastSyncErrorAt.Time
	}

	var p livechat.RecurrentCharge
	_ = json.Unmarshal(r.Payload, &p)

	subscription.Charge = &billing.Charge{
		ID:               r.ChargeID.String,
		LCOrganizationID: r.LcOrganizationID_2.String,
		Type:             billing.ChargeType(r.Type.String),
		Payload:          r.Payload,
		NextChargeAt:     p.NextChargeAt,
		CurrentChargeAt:  p.CurrentChargeAt,
		CreatedAt:        r.CreatedAt_2.Time,
		CanceledAt:       chargeDeletedAt,
		Status:           p.Status,
		SyncErrorCount:   int(r.SyncErrorCount.Int32),
		LastSyncErrorAt:  lastSyncErrorAt,
	}

	var dunningEndDate time.Time

	switch p.Status {
	case livechat.RecurrentChargeStatusPastDue, livechat.RecurrentChargeStatusFrozen:
		if p.NextChargeAt != nil {
			dunningEndDate = calculateNextDunningDate(*p.NextChargeAt)
		} else {
			dunningEndDate = calculateNextDunningDate(*p.CreatedAt)
		}
		subscription.DunningEndDate = &dunningEndDate
	case livechat.RecurrentChargeStatusActive:
		if p.NextChargeAt == nil {
			dunningEndDate = calculateNextDunningDate(*p.CreatedAt)
			subscription.DunningEndDate = &dunningEndDate
		}
	}

	return subscription
}

func calculateNextDunningDate(start time.Time) time.Time {
	dunningDate := start.AddDate(0, 0, 16)
	for dunningDate.Before(time.Now()) {
		dunningDate = dunningDate.AddDate(0, 0, 16)
	}
	return dunningDate
}
