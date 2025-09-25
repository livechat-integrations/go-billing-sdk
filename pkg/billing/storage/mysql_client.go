package storage

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time                         { return time.Now() }
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

type SQLCharge struct {
	ID               string     `json:"id" db:"id"`
	LcOrganizationID string     `json:"lc_organization_id" db:"lc_organization_id"`
	Type             string     `json:"type" db:"type"`
	Payload          string     `json:"payload" db:"payload"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	DeletedAt        *time.Time `json:"deleted_at" db:"deleted_at"`
}

type SQLSubscription struct {
	ID               string     `json:"id" db:"id"`
	LcOrganizationID string     `json:"lc_organization_id" db:"lc_organization_id"`
	PlanName         string     `json:"plan_name" db:"plan_name"`
	ChargeID         string     `json:"charge_id" db:"charge_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	DeletedAt        *time.Time `json:"deleted_at" db:"deleted_at"`
	Type             string     `json:"type" db:"type"`
	Payload          string     `json:"payload" db:"payload"`
	ChargeCreatedAt  time.Time  `json:"charge_created_at" db:"charge_created_at"`
	ChargeDeletedAt  *time.Time `json:"charge_deleted_at" db:"charge_deleted_at"`
}

// Make sure its Storage implementation
var _ billing.Storage = (*SQLClient)(nil)

type SQLClient struct {
	db    *sqlx.DB
	clock Clock
}

// NewSQLClient accepts a standard *sql.DB configured with github.com/go-sql-driver/mysql
func NewSQLClient(client *stdsql.DB, clock Clock) *SQLClient {
	return &SQLClient{
		db:    sqlx.NewDb(client, "mysql"),
		clock: clock,
	}
}

func (c *SQLClient) CreateCharge(ctx context.Context, ch billing.Charge) error {
	rawPayload, err := json.Marshal(ch.Payload)
	if err != nil {
		return err
	}

	res, err := c.db.ExecContext(ctx, "INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)", ch.ID, string(ch.Type), rawPayload, ch.LCOrganizationID, c.clock.Now())
	if err != nil {
		return fmt.Errorf("couldn't add new charge: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("couldn't add new charge")
	}

	return nil
}

func (c *SQLClient) GetCharge(ctx context.Context, id string) (*billing.Charge, error) {
	var ch SQLCharge
	if err := c.db.GetContext(ctx, &ch, "SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE id = ? AND deleted_at IS NULL", id); err != nil {
		if errors.Is(err, stdsql.ErrNoRows) {
			return nil, billing.ErrChargeNotFound
		}
		return nil, fmt.Errorf("couldn't select charge from DB: %w", err)
	}
	return ToBillingCharge(&ch), nil
}

func (c *SQLClient) UpdateChargePayload(ctx context.Context, id string, payload json.RawMessage) error {
	res, err := c.db.ExecContext(ctx, "UPDATE charges SET payload = ? WHERE id = ? AND deleted_at IS NULL", payload, id)
	if err != nil {
		return fmt.Errorf("couldn't update charge: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return billing.ErrChargeNotFound
	}
	return nil
}

func (c *SQLClient) DeleteCharge(ctx context.Context, id string) error {
	res, err := c.db.ExecContext(ctx, "UPDATE charges SET deleted_at = ? WHERE id = ?", c.clock.Now(), id)
	if err != nil {
		return fmt.Errorf("couldn't delete charge: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return billing.ErrChargeNotFound
	}

	return nil
}

func (c *SQLClient) CreateSubscription(ctx context.Context, subscription billing.Subscription) error {
	res, err := c.db.ExecContext(ctx, "INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at) VALUES (?, ?, ?, ?, ?)", subscription.ID, subscription.LCOrganizationID, subscription.PlanName, subscription.Charge.ID, c.clock.Now())
	if err != nil {
		return fmt.Errorf("couldn't add new subscription: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("couldn't add new subscription")
	}

	return nil
}

func (c *SQLClient) GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]billing.Subscription, error) {
	var subs []*SQLSubscription
	query := "SELECT s.id, s.lc_organization_id, s.plan_name, s.charge_id, s.created_at, s.deleted_at, c.type, c.payload, c.created_at AS charge_created_at, c.deleted_at AS charge_deleted_at FROM active_subscriptions s LEFT JOIN charges c ON s.charge_id = c.id AND s.lc_organization_id = c.lc_organization_id WHERE s.lc_organization_id = ?"
	if err := c.db.SelectContext(ctx, &subs, query, lcID); err != nil {
		return nil, fmt.Errorf("couldn't select subscriptions from DB: %w", err)
	}
	if len(subs) == 0 {
		return []billing.Subscription{}, nil
	}
	var subscriptions []billing.Subscription
	for _, sub := range subs {
		subscriptions = append(subscriptions, *ToBillingSubscription(sub))
	}
	return subscriptions, nil
}

func (c *SQLClient) DeleteSubscriptionByChargeID(ctx context.Context, lcID string, id string) error {
	res, err := c.db.ExecContext(ctx, "UPDATE subscriptions SET deleted_at = ? WHERE charge_id = ? AND lc_organization_id = ?", c.clock.Now(), id, lcID)
	if err != nil {
		return fmt.Errorf("couldn't delete subsctiption: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return billing.ErrSubscriptionNotFound
	}

	return nil
}

func (c *SQLClient) GetChargesByOrganizationID(ctx context.Context, lcID string) ([]billing.Charge, error) {
	var chs []*SQLCharge
	if err := c.db.SelectContext(ctx, &chs, "SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE lc_organization_id = ?", lcID); err != nil {
		return nil, fmt.Errorf("couldn't select charges from DB: %w", err)
	}
	if len(chs) == 0 {
		return []billing.Charge{}, nil
	}
	var charges []billing.Charge
	for _, ch := range chs {
		charges = append(charges, *ToBillingCharge(ch))
	}
	return charges, nil
}

func (c *SQLClient) CreateEvent(ctx context.Context, e events.Event) error {
	rawPayload, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}

	res, err := c.db.ExecContext(ctx, "INSERT INTO billing_events(id, lc_organization_id, type, action, payload, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)", e.ID, e.LCOrganizationID, string(e.Type), string(e.Action), rawPayload, e.Error, c.clock.Now())
	if err != nil {
		return fmt.Errorf("couldn't add new billing event: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("couldn't add new billing event")
	}

	return nil
}

func ToBillingSubscription(r *SQLSubscription) *billing.Subscription {
	var canceledAt *time.Time
	if r.DeletedAt != nil {
		canceledAt = r.DeletedAt
	}

	subscription := &billing.Subscription{
		ID:               r.ID,
		LCOrganizationID: r.LcOrganizationID,
		PlanName:         r.PlanName,
		CreatedAt:        r.CreatedAt,
		DeletedAt:        canceledAt,
	}

	if len(r.ChargeID) < 1 {
		return subscription
	}

	var chargeDeletedAt *time.Time
	if r.ChargeDeletedAt != nil {
		chargeDeletedAt = r.ChargeDeletedAt
	}

	subscription.Charge = &billing.Charge{
		ID:               r.ChargeID,
		LCOrganizationID: r.LcOrganizationID,
		Type:             billing.ChargeType(r.Type),
		Payload:          json.RawMessage(r.Payload),
		CreatedAt:        r.ChargeCreatedAt,
		CanceledAt:       chargeDeletedAt,
	}

	return subscription
}

func ToBillingCharge(c *SQLCharge) *billing.Charge {
	var canceledAt *time.Time
	if c.DeletedAt != nil {
		canceledAt = c.DeletedAt
	}

	return &billing.Charge{
		ID:               c.ID,
		LCOrganizationID: c.LcOrganizationID,
		Type:             billing.ChargeType(c.Type),
		Payload:          json.RawMessage(c.Payload),
		CreatedAt:        c.CreatedAt,
		CanceledAt:       canceledAt,
	}
}

// GetChargesByStatuses returns charges with JSON payload status in the provided list.
func (c *SQLClient) GetChargesByStatuses(ctx context.Context, statuses []string) ([]billing.Charge, error) {
	if len(statuses) == 0 {
		return []billing.Charge{}, nil
	}
	query, args, err := sqlx.In("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE JSON_UNQUOTE(JSON_EXTRACT(payload, '$.status')) IN (?) AND deleted_at IS NULL", statuses)
	if err != nil {
		return nil, fmt.Errorf("couldn't build query: %w", err)
	}
	// Ensure placeholders match driver; for mysql it's already '?'
	query = c.db.Rebind(query)
	var rows []*SQLCharge
	if err := c.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("couldn't select charges from DB: %w", err)
	}
	if len(rows) == 0 {
		return []billing.Charge{}, nil
	}
	var charges []billing.Charge
	for _, r := range rows {
		charges = append(charges, *ToBillingCharge(r))
	}
	return charges, nil
}

// DeleteSubscription marks subscription as deleted by its id and organization id.
func (c *SQLClient) DeleteSubscription(ctx context.Context, lcID, subID string) error {
	res, err := c.db.ExecContext(ctx, "UPDATE subscriptions SET deleted_at = ? WHERE id = ? AND lc_organization_id = ?", c.clock.Now(), subID, lcID)
	if err != nil {
		return fmt.Errorf("couldn't delete subsctiption: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return billing.ErrSubscriptionNotFound
	}
	return nil
}
