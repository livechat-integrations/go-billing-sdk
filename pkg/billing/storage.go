package billing

import (
	"context"
	"errors"
	"github.com/livechat-integrations/go-billing-sdk/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/pkg/events"
)

var (
	ErrChargeNotFound       = errors.New("charge not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Storage interface {
	CreateCharge(ctx context.Context, ic Charge) error
	GetCharge(ctx context.Context, id string) (*Charge, error)
	UpdateChargePayload(ctx context.Context, id string, payload livechat.BaseCharge) error
	DeleteCharge(ctx context.Context, id string) error
	GetChargesByOrganizationID(ctx context.Context, lcID string) ([]Charge, error)
	GetChargesByStatuses(ctx context.Context, statuses []string) ([]Charge, error)

	CreateSubscription(ctx context.Context, subscription Subscription) error
	GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]Subscription, error)
	DeleteSubscriptionByChargeID(ctx context.Context, lcID string, id string) error

	CreateEvent(ctx context.Context, event events.Event) error
}
