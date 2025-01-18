package billing

import "context"

type Storage interface {
	CreateCharge(ctx context.Context, ic Charge) error
	GetCharge(ctx context.Context, id string) (*Charge, error)
	UpdateChargePayload(ctx context.Context, id string, payload BaseCharge) error
	DeleteCharge(ctx context.Context, id string) error

	CreateSubscription(ctx context.Context, subscription Subscription) error
	GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]Subscription, error)
	DeleteSubscriptionByChargeID(ctx context.Context, id string) error
}
