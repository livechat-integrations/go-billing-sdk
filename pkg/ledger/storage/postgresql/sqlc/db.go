//go:generate go run -mod=readonly github.com/sqlc-dev/sqlc/cmd/sqlc@v1.25.0 -x -f ./sqlc.yaml generate

package sqlc

import (
	"github.com/livechat-integrations/go-billing-sdk/pkg/ledger"
)

func (o *LedgerLedger) ToLedgerOperation() (*ledger.Operation, error) {
	v, err := o.Amount.Float64Value()
	if err != nil {
		return nil, err
	}

	return &ledger.Operation{
		ID:               o.ID,
		LCOrganizationID: o.LcOrganizationID,
		Amount:           float32(v.Float64),
		Payload:          o.Payload,
		CreatedAt:        o.CreatedAt.Time,
	}, nil
}

func (t *LedgerTopUp) ToLedgerTopUp() (*ledger.TopUp, error) {
	v, err := t.Amount.Float64Value()
	if err != nil {
		return nil, err
	}

	tu := &ledger.TopUp{
		ID:               t.ID,
		LCOrganizationID: t.LcOrganizationID,
		Status:           ledger.TopUpStatus(t.Status),
		Amount:           float32(v.Float64),
		Type:             ledger.TopUpType(t.Type),
		ConfirmationUrl:  t.ConfirmationUrl,
		LCCharge:         t.LcCharge,
		CreatedAt:        t.CreatedAt.Time,
		UpdatedAt:        t.UpdatedAt.Time,
	}

	if t.NextTopUpAt.Valid {
		tu.NextTopUpAt = &t.NextTopUpAt.Time
	}
	if t.CurrentToppedUpAt.Valid {
		tu.CurrentToppedUpAt = &t.CurrentToppedUpAt.Time
	}

	return tu, nil
}
