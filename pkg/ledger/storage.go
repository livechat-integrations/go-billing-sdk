package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/pkg/events"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type InitRecurrentTopUpRequiredValuesParams struct {
	CurrentToppedUpAt time.Time
	NextTopUpAt       time.Time
	ID                string
}

type UpdateTopUpStatusParams struct {
	ID                string
	Status            TopUpStatus
	CurrentToppedUpAt *time.Time
}

type GetTopUpByIDAndTypeParams struct {
	ID                string
	Type              TopUpType
	CurrentToppedUpAt *time.Time
}

type Storage interface {
	CreateCharge(ctx context.Context, charge Charge) error
	UpdateChargeStatus(ctx context.Context, ID string, status ChargeStatus) error
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]TopUp, error)
	UpdateTopUpStatus(ctx context.Context, params UpdateTopUpStatusParams) error
	GetTopUpByIDAndType(ctx context.Context, params GetTopUpByIDAndTypeParams) (*TopUp, error)
	CreateEvent(ctx context.Context, event events.Event) error
	GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error)
	UpsertTopUp(ctx context.Context, topUp TopUp) (*TopUp, error)
	InitRecurrentTopUpRequiredValues(ctx context.Context, params InitRecurrentTopUpRequiredValuesParams) error
}
