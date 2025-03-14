package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/livechat-integrations/go-billing-sdk/pkg/ledger"
	"github.com/livechat-integrations/go-billing-sdk/pkg/ledger/storage/postgresql/sqlc"
)

type PGXConn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type PostgresqlPGX struct {
	queries *sqlc.Queries
}

func NewPostgresqlPGX(conn PGXConn) *PostgresqlPGX {
	return &PostgresqlPGX{queries: sqlc.New(conn)}
}

func (r *PostgresqlPGX) CreateCharge(ctx context.Context, c ledger.Charge) error {
	rawLCCharge, err := json.Marshal(c.LCCharge)
	if err != nil {
		return err
	}

	if err = r.queries.CreateCharge(ctx, sqlc.CreateChargeParams{
		ID:               c.ID,
		Amount:           ToPGNumeric(&c.Amount),
		Type:             string(c.Type),
		Status:           string(c.Status),
		LcCharge:         rawLCCharge,
		LcOrganizationID: c.LCOrganizationID,
	}); err != nil {
		return err
	}

	return nil
}

func (r *PostgresqlPGX) UpdateChargeStatus(ctx context.Context, ID string, status ledger.ChargeStatus) error {
	err := r.queries.UpdateChargeStatus(ctx, sqlc.UpdateChargeStatusParams{
		ID:     ID,
		Status: string(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetChargeByIdAndType(ctx context.Context, ID string, chargeType ledger.ChargeType) (*ledger.Charge, error) {
	t, err := r.queries.GetChargeByIDAndTypeWhereStatusIsNot(ctx, sqlc.GetChargeByIDAndTypeWhereStatusIsNotParams{
		ID:     ID,
		Type:   string(chargeType),
		Status: string(ledger.ChargeStatusCancelled),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return t.ToLedgerCharge()
}

func (r *PostgresqlPGX) CreateTopUp(ctx context.Context, t ledger.TopUp) error {
	rawLCCharge, err := json.Marshal(t.LCCharge)
	if err != nil {
		return err
	}
	if err := r.queries.CreateTopUp(ctx, sqlc.CreateTopUpParams{
		ID:               t.ID,
		Status:           string(t.Status),
		Amount:           ToPGNumeric(&t.Amount),
		Type:             string(t.Type),
		LcOrganizationID: t.LCOrganizationID,
		LcCharge:         rawLCCharge,
		ConfirmationUrl:  t.ConfirmationUrl,
		CurrentToppedUpAt: pgtype.Timestamptz{
			Time:  t.CurrentToppedUpAt,
			Valid: true,
		},
		NextTopUpAt: pgtype.Timestamptz{
			Time:  t.NextTopUpAt,
			Valid: true,
		},
	}); err != nil {
		return err
	}
	return nil
}

func (r *PostgresqlPGX) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	b, err := r.queries.GetOrganizationBalance(ctx, sqlc.GetOrganizationBalanceParams{
		LcOrganizationID: organizationID,
		Status:           string(ledger.TopUpStatusActive),
		Status_2:         string(ledger.ChargeStatusActive),
	})
	if err != nil {
		return 0, err
	}
	v, err := b.Float64Value()
	if err != nil {
		return 0, err
	}

	return float32(v.Float64), nil
}

func (r *PostgresqlPGX) GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]ledger.TopUp, error) {
	var ts []ledger.TopUp
	rows, err := r.queries.GetTopUpsByOrganizationID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ts, nil
		}
		return nil, err
	}
	for _, row := range rows {
		t, err := row.ToLedgerTopUp()
		if err != nil {
			return nil, err
		}
		ts = append(ts, *t)
	}
	return ts, nil
}

func (r *PostgresqlPGX) UpdateTopUpStatus(ctx context.Context, ID string, status ledger.TopUpStatus) error {
	err := r.queries.UpdateTopUpRequestStatus(ctx, sqlc.UpdateTopUpRequestStatusParams{
		ID:     ID,
		Status: string(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetTopUpByIdAndType(ctx context.Context, ID string, topUpType ledger.TopUpType) (*ledger.TopUp, error) {
	t, err := r.queries.GetTopUpByIDAndTypeWhereStatusIsNot(ctx, sqlc.GetTopUpByIDAndTypeWhereStatusIsNotParams{
		ID:     ID,
		Type:   string(topUpType),
		Status: string(ledger.TopUpStatusCancelled),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return t.ToLedgerTopUp()
}

func (r *PostgresqlPGX) CreateEvent(ctx context.Context, e ledger.Event) error {
	err := r.queries.CreateEvent(ctx, sqlc.CreateEventParams{
		ID:               e.ID,
		LcOrganizationID: e.LCOrganizationID,
		Type:             string(e.Type),
		Action:           string(e.Action),
		Payload:          e.Payload,
	})
	if err != nil {
		return err
	}
	return nil
}

func ToPGNumeric(n *float32) pgtype.Numeric {
	if n == nil {
		return pgtype.Numeric{}
	}

	v := pgtype.Numeric{}
	_ = v.Scan(fmt.Sprintf("%f", *n))

	return v
}
