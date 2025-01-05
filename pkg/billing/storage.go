package billing

import "context"

type storage interface {
	CreateCharge(ctx context.Context, ic InstallationCharge) error
	GetCharge(ctx context.Context, id string) (*InstallationCharge, error)
	UpdateChargePayload(ctx context.Context, id string, payload BaseCharge) error
	GetChargeByInstallationID(ctx context.Context, lcID string) (*InstallationCharge, error)
}
