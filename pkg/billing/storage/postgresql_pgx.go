package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing/storage/postgresql/sqlc"
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

func (r *PostgresqlPGX) CreateCharge(ctx context.Context, c billing.Charge) error {
	rawPayload, err := json.Marshal(c.Payload)
	if err != nil {
		return err
	}

	if err = r.queries.CreateCharge(ctx, sqlc.CreateChargeParams{
		ID:               c.ID,
		Type:             string(c.Type),
		LcOrganizationID: c.LCOrganizationID,
		Payload:          rawPayload,
	}); err != nil {
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetCharge(ctx context.Context, id string) (*billing.Charge, error) {
	row, err := r.queries.GetChargeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return row.ToBillingCharge(), nil
}

func (r *PostgresqlPGX) GetChargeByOrganizationID(ctx context.Context, lcID string) (*billing.Charge, error) {
	row, err := r.queries.GetChargeByOrganizationID(ctx, lcID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return row.ToBillingCharge(), nil
}

func (r *PostgresqlPGX) UpdateChargePayload(ctx context.Context, id string, payload billing.BaseCharge) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.queries.UpdateCharge(ctx, sqlc.UpdateChargeParams{
		ID:      id,
		Payload: rawPayload,
	})
}

func (r *PostgresqlPGX) CreateSubscription(ctx context.Context, subscription billing.Subscription) error {
	if err := r.queries.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
		ID:               subscription.ID,
		LcOrganizationID: subscription.LCOrganizationID,
		PlanName:         subscription.PlanName,
		ChargeID:         pgtype.Text{String: subscription.Charge.ID, Valid: true},
	}); err != nil {
		return err
	}

	return nil
}

func (r *PostgresqlPGX) GetSubscriptionByOrganizationID(ctx context.Context, lcID string) (*billing.Subscription, error) {
	row, err := r.queries.GetSubscriptionByOrganizationID(ctx, lcID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return row.ToBillingSubscription(), nil
}
