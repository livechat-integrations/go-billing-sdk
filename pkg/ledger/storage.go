package ledger

import (
	"context"
	"fmt"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type Storage interface {
	GetTopUpByID(ctx context.Context, ID string) (*TopUp, error)
	GetTopUpByIDAndType(ctx context.Context, ID string, topUpType TopUpType) (*TopUp, error)
	GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]TopUp, error)
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	CreateCharge(ctx context.Context, charge Charge) error
	CreateTopUp(ctx context.Context, topUp TopUp) error
	CreateEvent(ctx context.Context, event Event) error
	UpdateTopUpStatus(ctx context.Context, ID string, status TopUpStatus) error
	UpdateChargeStatus(ctx context.Context, ID string, status ChargeStatus) error
	GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error)
	UpsertTopUp(ctx context.Context, topUp TopUp) (*TopUp, error)
}
