package billing

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

var (
	ErrChargeNotFound       = errors.New("charge not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Storage interface {
	CreateCharge(ctx context.Context, ic Charge) error
	GetCharge(ctx context.Context, id string) (*Charge, error)
	UpdateChargePayload(ctx context.Context, id string, payload json.RawMessage) error
	DeleteCharge(ctx context.Context, id string) error
	GetChargesByOrganizationID(ctx context.Context, lcID string) ([]Charge, error)
	GetChargesByStatuses(ctx context.Context, statuses []string) ([]Charge, error)
	IncrementChargeSyncErrorCount(ctx context.Context, chargeID string) error
	GetChargesWithHighErrorCount(ctx context.Context, threshold int) ([]Charge, error)

	CreateSubscription(ctx context.Context, subscription Subscription) error
	GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]Subscription, error)
	DeleteSubscriptionByChargeID(ctx context.Context, lcID string, id string) error
	DeleteSubscription(ctx context.Context, lcID, subID string) error

	CreateEvent(ctx context.Context, event events.Event) error

	// Trial management
	RecordTrialUsage(ctx context.Context, lcOrganizationID string) error
	HasUsedTrial(ctx context.Context, lcOrganizationID string) (bool, error)
}
