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
	ID     string
	Status TopUpStatus
}

type GetTopUpByIDAndTypeParams struct {
	ID   string
	Type TopUpType
}

type GetTopUpsByTypeWhereStatusNotInParams struct {
	Type     TopUpType
	Statuses []TopUpStatus
}

type Storage interface {
	CreateLedgerOperation(ctx context.Context, c Operation) error
	GetLedgerOperations(ctx context.Context, organizationID string) ([]Operation, error)
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]TopUp, error)
	GetTopUpByIDAndOrganizationID(ctx context.Context, organizationID string, id string) (*TopUp, error)
	GetTopUpsByTypeWhereStatusNotIn(ctx context.Context, params GetTopUpsByTypeWhereStatusNotInParams) ([]TopUp, error)
	GetRecurrentTopUpsWhereStatusNotIn(ctx context.Context, statuses []TopUpStatus) ([]TopUp, error)
	GetDirectTopUpsWithoutOperations(ctx context.Context) ([]TopUp, error)
	UpdateTopUpStatus(ctx context.Context, params UpdateTopUpStatusParams) error
	GetTopUpByIDAndType(ctx context.Context, params GetTopUpByIDAndTypeParams) (*TopUp, error)
	CreateEvent(ctx context.Context, event events.Event) error
	GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error)
	UpsertTopUp(ctx context.Context, topUp TopUp) (*TopUp, error)
}
