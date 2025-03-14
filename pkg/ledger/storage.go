package ledger

import "context"

type Storage interface {
	GetChargeByIdAndType(ctx context.Context, ID string, chargeType ChargeType) (*Charge, error)
	GetTopUpByIdAndType(ctx context.Context, ID string, topUpType TopUpType) (*TopUp, error)
	GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]TopUp, error)
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	CreateCharge(ctx context.Context, ic Charge) error
	CreateTopUp(ctx context.Context, it TopUp) error
	CreateEvent(ctx context.Context, e Event) error
	UpdateTopUpStatus(ctx context.Context, ID string, status TopUpStatus) error
	UpdateChargeStatus(ctx context.Context, ID string, status ChargeStatus) error
}
