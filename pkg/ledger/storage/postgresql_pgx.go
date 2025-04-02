package storage

import (
	"context"
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
	if err := r.queries.CreateCharge(ctx, sqlc.CreateChargeParams{
		ID:               c.ID,
		Amount:           ToPGNumeric(&c.Amount),
		Status:           string(c.Status),
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
			return ledger.ErrNotFound
		}
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	b, err := r.queries.GetOrganizationBalance(ctx, sqlc.GetOrganizationBalanceParams{
		LcOrganizationID: organizationID,
		Status:           string(ledger.TopUpStatusActive),
		Status_2:         string(ledger.ChargeStatusCancelled),
		Status_3:         string(ledger.ChargeStatusActive),
	})
	if err != nil {
		return float32(0), err
	}
	v, err := b.Float64Value()
	if err != nil {
		return float32(0), err
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

func (r *PostgresqlPGX) InitRecurrentTopUpRequiredValues(ctx context.Context, params ledger.InitRecurrentTopUpRequiredValuesParams) error {
	p := sqlc.InitTopUpRequiredValuesParams{
		CurrentToppedUpAt: pgtype.Timestamptz{
			Time:  params.CurrentToppedUpAt,
			Valid: true,
		},
		NextTopUpAt: pgtype.Timestamptz{
			Time:  params.NextTopUpAt,
			Valid: true,
		},
		UniqueAt: pgtype.Timestamptz{
			Time:  params.CurrentToppedUpAt,
			Valid: true,
		},
		ID:     params.ID,
		Type:   string(ledger.TopUpTypeRecurrent),
		Status: string(ledger.TopUpStatusPending),
	}

	err := r.queries.InitTopUpRequiredValues(ctx, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *PostgresqlPGX) UpdateTopUpStatus(ctx context.Context, params ledger.UpdateTopUpStatusParams) error {
	p := sqlc.UpdateTopUpRequestStatusParams{
		ID:       params.ID,
		Status:   string(params.Status),
		UniqueAt: pgtype.Timestamptz{},
	}
	if params.CurrentToppedUpAt != nil {
		p.UniqueAt = pgtype.Timestamptz{
			Time:  *params.CurrentToppedUpAt,
			Valid: true,
		}
	}

	err := r.queries.UpdateTopUpRequestStatus(ctx, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ledger.ErrNotFound
		}
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetTopUpByIDAndType(ctx context.Context, params ledger.GetTopUpByIDAndTypeParams) (*ledger.TopUp, error) {
	p := sqlc.GetTopUpByIDAndTypeWhereStatusIsNotParams{
		ID:     params.ID,
		Type:   string(params.Type),
		Status: string(ledger.TopUpStatusCancelled),
	}

	t, err := r.queries.GetTopUpByIDAndTypeWhereStatusIsNot(ctx, p)
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
		Error: pgtype.Text{
			String: e.Error,
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresqlPGX) GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status ledger.TopUpStatus) ([]ledger.TopUp, error) {
	rts, err := r.queries.GetTopUpsByOrganizationIDAndStatus(ctx, sqlc.GetTopUpsByOrganizationIDAndStatusParams{
		LcOrganizationID: organizationID,
		Status:           string(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []ledger.TopUp{}, nil
		}
		return nil, err
	}

	var topUps []ledger.TopUp
	for _, rt := range rts {
		topUp, err := rt.ToLedgerTopUp()
		if err != nil {
			return nil, err
		}
		topUps = append(topUps, *topUp)
	}

	return topUps, nil
}

func (r *PostgresqlPGX) UpsertTopUp(ctx context.Context, topUp ledger.TopUp) (*ledger.TopUp, error) {
	params := sqlc.UpsertTopUpParams{
		ID:               topUp.ID,
		Status:           string(topUp.Status),
		Amount:           ToPGNumeric(&topUp.Amount),
		Type:             string(topUp.Type),
		LcOrganizationID: topUp.LCOrganizationID,
		LcCharge:         topUp.LCCharge,
		ConfirmationUrl:  topUp.ConfirmationUrl,
	}

	if topUp.CurrentToppedUpAt != nil {
		params.CurrentToppedUpAt = pgtype.Timestamptz{
			Time:  *topUp.CurrentToppedUpAt,
			Valid: true,
		}
		params.Column9 = pgtype.Timestamptz{
			Time:  *topUp.CurrentToppedUpAt,
			Valid: true,
		}
	}
	if topUp.NextTopUpAt != nil {
		params.NextTopUpAt = pgtype.Timestamptz{
			Time:  *topUp.NextTopUpAt,
			Valid: true,
		}
	}

	t, err := r.queries.UpsertTopUp(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ledger.ErrNotFound
		}
		return nil, err
	}

	return t.ToLedgerTopUp()
}

func ToPGNumeric(n *float32) pgtype.Numeric {
	if n == nil {
		return pgtype.Numeric{}
	}

	v := pgtype.Numeric{}
	_ = v.Scan(fmt.Sprintf("%f", *n))

	return v
}
