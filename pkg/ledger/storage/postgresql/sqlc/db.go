//go:generate go run -mod=readonly github.com/sqlc-dev/sqlc/cmd/sqlc@v1.25.0 -x -f ./sqlc.yaml generate

package sqlc

import (
	"github.com/livechat-integrations/go-billing-sdk/pkg/ledger"
)

func (c *Charge) ToLedgerCharge() (*ledger.Charge, error) {
	v, err := c.Amount.Float64Value()
	if err != nil {
		return nil, err
	}

	return &ledger.Charge{
		ID:               c.ID,
		LCOrganizationID: c.LcOrganizationID,
		Amount:           float32(v.Float64),
		Status:           ledger.ChargeStatus(c.Status),
		CreatedAt:        c.CreatedAt.Time,
		UpdatedAt:        c.UpdatedAt.Time,
	}, nil
}

func (t *TopUp) ToLedgerTopUp() (*ledger.TopUp, error) {
	v, err := t.Amount.Float64Value()
	if err != nil {
		return nil, err
	}

	return &ledger.TopUp{
		ID:                t.ID,
		LCOrganizationID:  t.LcOrganizationID,
		Status:            ledger.TopUpStatus(t.Status),
		Amount:            float32(v.Float64),
		Type:              ledger.TopUpType(t.Type),
		ConfirmationUrl:   t.ConfirmationUrl,
		CurrentToppedUpAt: t.CurrentToppedUpAt.Time,
		NextTopUpAt:       t.NextTopUpAt.Time,
		LCCharge:          t.LcCharge,
		CreatedAt:         t.CreatedAt.Time,
		UpdatedAt:         t.UpdatedAt.Time,
	}, nil
}
