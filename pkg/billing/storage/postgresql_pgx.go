package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/billing/storage/postgresql/sqlc"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
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

func (r *PostgresqlPGX) UpdateChargePayload(ctx context.Context, id string, payload json.RawMessage) error {
	return r.queries.UpdateCharge(ctx, sqlc.UpdateChargeParams{
		ID:      id,
		Payload: payload,
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

func (r *PostgresqlPGX) GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]billing.Subscription, error) {
	rows, err := r.queries.GetSubscriptionsByOrganizationID(ctx, lcID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var subscriptions []billing.Subscription
	for _, row := range rows {
		subscriptions = append(subscriptions, *row.ToBillingSubscription())
	}
	return subscriptions, nil
}

func (r *PostgresqlPGX) DeleteCharge(ctx context.Context, id string) error {
	return r.queries.DeleteCharge(ctx, id)
}

func (r *PostgresqlPGX) DeleteSubscriptionByChargeID(ctx context.Context, lcID string, id string) error {
	err := r.queries.DeleteSubscriptionByChargeID(ctx, sqlc.DeleteSubscriptionByChargeIDParams{
		ChargeID:         pgtype.Text{String: id, Valid: true},
		LcOrganizationID: lcID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return billing.ErrSubscriptionNotFound
	}

	return err
}

func (r *PostgresqlPGX) GetChargesByOrganizationID(ctx context.Context, lcID string) ([]billing.Charge, error) {
	rows, err := r.queries.GetChargesByOrganizationID(ctx, lcID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var charges []billing.Charge
	for _, row := range rows {
		charges = append(charges, *row.ToBillingCharge())
	}
	return charges, nil
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

func (r *PostgresqlPGX) GetChargesByStatuses(ctx context.Context, statuses []string) ([]billing.Charge, error) {
	rows, err := r.queries.GetChargesByStatuses(ctx, statuses)
	if err != nil {
		return nil, err
	}

	var charges []billing.Charge
	for _, row := range rows {
		charges = append(charges, *row.ToBillingCharge())
	}

	return charges, nil
}

func (r *PostgresqlPGX) DeleteSubscription(ctx context.Context, lcID, subID string) error {
	return r.queries.DeleteSubscription(ctx, sqlc.DeleteSubscriptionParams{
		ID:               subID,
		LcOrganizationID: lcID,
	})
}
