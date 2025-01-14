package billing

import "context"

type Storage interface {
	CreateCharge(ctx context.Context, ic Charge) error
	GetCharge(ctx context.Context, id string) (*Charge, error)
	UpdateChargePayload(ctx context.Context, id string, payload BaseCharge) error
	GetChargeByOrganizationID(ctx context.Context, lcID string) (*Charge, error)

	CreateSubscription(ctx context.Context, subscription Subscription) error
	GetSubscriptionByOrganizationID(ctx context.Context, lcID string) (*Subscription, error)
}
