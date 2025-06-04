package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/ledger"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/ledger/storage/postgresql/sqlc"
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

func (r *PostgresqlPGX) CreateLedgerOperation(ctx context.Context, c ledger.Operation) error {
	if err := r.queries.CreateLedgerOperation(ctx, sqlc.CreateLedgerOperationParams{
		ID:               c.ID,
		Amount:           ToPGNumeric(&c.Amount),
		LcOrganizationID: c.LCOrganizationID,
		Payload:          c.Payload,
	}); err != nil {
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetLedgerOperations(ctx context.Context, organizationID string) ([]ledger.Operation, error) {
	var ops []ledger.Operation
	rows, err := r.queries.GetLedgerOperationsByOrganizationID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ops, nil
		}
		return nil, err
	}
	for _, row := range rows {
		o, err := row.ToLedgerOperation()
		if err != nil {
			return nil, err
		}
		ops = append(ops, *o)
	}
	return ops, nil
}

func (r *PostgresqlPGX) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	b, err := r.queries.GetOrganizationBalance(ctx, organizationID)
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
	dbTopUps, err := r.queries.GetTopUpsByOrganizationID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ts, nil
		}
		return nil, err
	}
	return ToTopUps(dbTopUps)
}

func (r *PostgresqlPGX) UpdateTopUpStatus(ctx context.Context, params ledger.UpdateTopUpStatusParams) error {
	p := sqlc.UpdateTopUpRequestStatusParams{
		ID:     params.ID,
		Status: string(params.Status),
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

func (r *PostgresqlPGX) CreateEvent(ctx context.Context, e events.Event) error {
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
	dbTopUps, err := r.queries.GetTopUpsByOrganizationIDAndStatus(ctx, sqlc.GetTopUpsByOrganizationIDAndStatusParams{
		LcOrganizationID: organizationID,
		Status:           string(status),
	})
	if err != nil {
		return HandleTopUpsError(err)
	}
	return ToTopUps(dbTopUps)
}

func (r *PostgresqlPGX) GetTopUpByIDAndOrganizationID(ctx context.Context, organizationID string, id string) (*ledger.TopUp, error) {
	topUp, err := r.queries.GetTopUpByIDAndOrganizationID(ctx, sqlc.GetTopUpByIDAndOrganizationIDParams{
		LcOrganizationID: organizationID,
		ID:               id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return topUp.ToLedgerTopUp()
}

func (r *PostgresqlPGX) GetTopUpsByTypeWhereStatusNotIn(ctx context.Context, params ledger.GetTopUpsByTypeWhereStatusNotInParams) ([]ledger.TopUp, error) {
	var statuses []string
	for _, s := range params.Statuses {
		statuses = append(statuses, string(s))
	}
	dbTopUps, err := r.queries.GetTopUpsByTypeWhereStatusNotIn(ctx, sqlc.GetTopUpsByTypeWhereStatusNotInParams{
		Type:    string(params.Type),
		Column2: statuses,
	})
	if err != nil {
		return HandleTopUpsError(err)
	}
	return ToTopUps(dbTopUps)
}

func (r *PostgresqlPGX) GetRecurrentTopUpsWhereStatusNotIn(ctx context.Context, statuses []ledger.TopUpStatus) ([]ledger.TopUp, error) {
	var stringStatuses []string
	for _, s := range statuses {
		stringStatuses = append(stringStatuses, string(s))
	}
	dbTopUps, err := r.queries.GetRecurrentTopUpsWhereStatusNotIn(ctx, stringStatuses)
	if err != nil {
		return HandleTopUpsError(err)
	}
	return ToTopUps(dbTopUps)
}

func (r *PostgresqlPGX) GetDirectTopUpsWithoutOperations(ctx context.Context) ([]ledger.TopUp, error) {
	dbTopUps, err := r.queries.GetDirectTopUpsWithoutOperations(ctx)
	if err != nil {
		return HandleTopUpsError(err)
	}
	return ToTopUps(dbTopUps)
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

func ToTopUps(dbTopUps []sqlc.LedgerTopUp) ([]ledger.TopUp, error) {
	var topUps []ledger.TopUp
	for _, rt := range dbTopUps {
		topUp, err := rt.ToLedgerTopUp()
		if err != nil {
			return nil, err
		}
		topUps = append(topUps, *topUp)
	}
	return topUps, nil
}

func HandleTopUpsError(err error) ([]ledger.TopUp, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return []ledger.TopUp{}, nil
	}
	return nil, err
}
