package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	lcMySQL "github.com/livechat/go-mysql"
	"github.com/rcrowley/go-metrics"

	"github.com/livechat-integrations/go-billing-sdk/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/pkg/events"
)

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time                         { return time.Now() }
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

const slowlogTreshold = time.Second

type MySQLClient interface {
	Exec(ctx context.Context, query string, args ...interface{}) (*lcMySQL.Meta, error)
	Query(ctx context.Context, query string, args ...interface{}) (*lcMySQL.Results, error)
	SamplesChan() chan *lcMySQL.QueryStats
}

type SQLCharge struct {
	ID               string     `json:"id" mysql:"id"`
	LcOrganizationID string     `json:"lc_organization_id" mysql:"lc_organization_id"`
	Type             string     `json:"type" mysql:"type"`
	Payload          string     `json:"payload" mysql:"payload"`
	CreatedAt        time.Time  `json:"created_at" mysql:"created_at"`
	DeletedAt        *time.Time `json:"deleted_at" mysql:"deleted_at"`
}

type SQLSubscription struct {
	ID               string     `json:"id" mysql:"id"`
	LcOrganizationID string     `json:"lc_organization_id" mysql:"lc_organization_id"`
	PlanName         string     `json:"plan_name" mysql:"plan_name"`
	ChargeID         string     `json:"charge_id" mysql:"charge_id"`
	CreatedAt        time.Time  `json:"created_at" mysql:"created_at"`
	DeletedAt        *time.Time `json:"deleted_at" mysql:"deleted_at"`
	Type             string     `json:"type" mysql:"type"`
	Payload          string     `json:"payload" mysql:"payload"`
	ChargeCreatedAt  time.Time  `json:"charge_created_at" mysql:"charge_created_at"`
	ChargeDeletedAt  *time.Time `json:"charge_deleted_at" mysql:"charge_deleted_at"`
}

type SQLClient struct {
	sqlClient MySQLClient
	clock     Clock
}

func NewSQLClient(client MySQLClient, clock Clock) *SQLClient {
	return &SQLClient{
		sqlClient: client,
		clock:     clock,
	}
}

func (sql *SQLClient) CreateCharge(ctx context.Context, c billing.Charge) error {
	rawPayload, err := json.Marshal(c.Payload)
	if err != nil {
		return err
	}

	res, err := sql.sqlClient.Exec(ctx, "INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)", c.ID, string(c.Type), rawPayload, c.LCOrganizationID, sql.clock.Now())
	if err != nil {
		return fmt.Errorf("couldn't add new charge: %w", err)
	}
	if res.RowsAffected == 0 {
		return errors.New("couldn't add new charge")
	}

	return nil
}

func (sql *SQLClient) GetCharge(ctx context.Context, id string) (*billing.Charge, error) {
	res, err := sql.sqlClient.Query(ctx, "SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return nil, fmt.Errorf("couldn't select charge from DB: %w", err)
	}
	if res.Count() != 1 {
		return nil, billing.ErrChargeNotFound
	}
	var charges []*SQLCharge
	if err := res.CastTo(&charges); err != nil {
		return nil, fmt.Errorf("couldn't cast result to sql charge: %w", err)
	}
	if charges[0] == nil {
		return nil, billing.ErrChargeNotFound
	}
	return ToBillingCharge(charges[0]), nil
}

func (sql *SQLClient) UpdateChargePayload(ctx context.Context, id string, payload livechat.BaseCharge) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := sql.sqlClient.Exec(ctx, "UPDATE charges SET payload = ? WHERE id = ? AND deleted_at IS NULL", rawPayload, id)
	if err != nil {
		return fmt.Errorf("couldn't update charge: %w", err)
	}
	if res.RowsAffected == 0 {
		return billing.ErrChargeNotFound
	}

	return nil
}

func (sql *SQLClient) DeleteCharge(ctx context.Context, id string) error {
	res, err := sql.sqlClient.Exec(ctx, "UPDATE charges SET deleted_at = ? WHERE id = ?", sql.clock.Now(), id)
	if err != nil {
		return fmt.Errorf("couldn't delete charge: %w", err)
	}
	if res.RowsAffected == 0 {
		return billing.ErrChargeNotFound
	}

	return nil
}

func (sql *SQLClient) CreateSubscription(ctx context.Context, subscription billing.Subscription) error {
	res, err := sql.sqlClient.Exec(ctx, "INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at) VALUES (?, ?, ?, ?, ?)", subscription.ID, subscription.LCOrganizationID, subscription.PlanName, subscription.Charge.ID, sql.clock.Now())
	if err != nil {
		return fmt.Errorf("couldn't add new subscription: %w", err)
	}
	if res.RowsAffected == 0 {
		return errors.New("couldn't add new subscription")
	}

	return nil
}

func (sql *SQLClient) GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]billing.Subscription, error) {
	res, err := sql.sqlClient.Query(ctx, "SELECT s.id, s.lc_organization_id, s.plan_name, s.charge_id, s.created_at, s.deleted_at, c.type, c.payload, c.created_at AS charge_created_at, c.deleted_at AS charge_deleted_at FROM active_subscriptions s LEFT JOIN charges c ON s.charge_id = c.id AND s.lc_organization_id = c.lc_organization_id WHERE s.lc_organization_id = ?", lcID)
	if err != nil {
		return nil, fmt.Errorf("couldn't select subscriptions from DB: %w", err)
	}
	if res.Count() < 1 {
		return []billing.Subscription{}, nil
	}
	var subs []*SQLSubscription
	if err := res.CastTo(&subs); err != nil {
		return nil, fmt.Errorf("couldn't cast result to sql subscriptions: %w", err)
	}

	var subscriptions []billing.Subscription
	for _, sub := range subs {
		subscriptions = append(subscriptions, *ToBillingSubscription(sub))
	}
	return subscriptions, nil
}

func (sql *SQLClient) DeleteSubscriptionByChargeID(ctx context.Context, lcID string, id string) error {
	res, err := sql.sqlClient.Exec(ctx, "UPDATE subscriptions SET deleted_at = ? WHERE charge_id = ? AND lc_organization_id = ?", sql.clock.Now(), id, lcID)
	if err != nil {
		return fmt.Errorf("couldn't delete subsctiption: %w", err)
	}
	if res.RowsAffected == 0 {
		return billing.ErrSubscriptionNotFound
	}

	return nil
}

func (sql *SQLClient) GetChargesByOrganizationID(ctx context.Context, lcID string) ([]billing.Charge, error) {
	res, err := sql.sqlClient.Query(ctx, "SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE lc_organization_id = ?", lcID)
	if err != nil {
		return nil, fmt.Errorf("couldn't select charges from DB: %w", err)
	}
	if res.Count() < 1 {
		return []billing.Charge{}, nil
	}
	var chs []*SQLCharge
	if err := res.CastTo(&chs); err != nil {
		return nil, fmt.Errorf("couldn't cast result to sql charges: %w", err)
	}

	var charges []billing.Charge
	for _, ch := range chs {
		charges = append(charges, *ToBillingCharge(ch))
	}
	return charges, nil
}

func (sql *SQLClient) CreateEvent(ctx context.Context, e events.Event) error {
	rawPayload, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}

	res, err := sql.sqlClient.Exec(ctx, "INSERT INTO billing_events(id, lc_organization_id, type, action, payload, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)", e.ID, e.LCOrganizationID, string(e.Type), string(e.Action), rawPayload, e.Error, sql.clock.Now())
	if err != nil {
		return fmt.Errorf("couldn't add new billing event: %w", err)
	}
	if res.RowsAffected == 0 {
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

func (sql *SQLClient) Metrics() {
	for {
		sample := <-sql.sqlClient.SamplesChan()
		metrics.GetOrRegisterTimer("db.execution_time", nil).Update(sample.ExecutionTime)
		if sample.ExecutionTime > slowlogTreshold {
			slog.Warn("slowlog for query '%s': %v", sample.Query, sample.ExecutionTime)
		}
	}
}
