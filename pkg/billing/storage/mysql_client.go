package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	lcMySQL "github.com/livechat/go-mysql"

	"github.com/livechat-integrations/go-billing-sdk/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

var (
	ErrChargeNotFound       = errors.New("charge not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type MySQLClient interface {
	Exec(ctx context.Context, query string, args ...interface{}) (*lcMySQL.Meta, error)
	Query(ctx context.Context, query string, args ...interface{}) (*lcMySQL.Results, error)
	SamplesChan() chan *lcMySQL.QueryStats
}

type SQLCharge struct {
	ID               string     `json:"id" mysql:"id"`
	LcOrganizationID string     `json:"lc_organization_id" mysql:"lc_organization_id"`
	Type             string     `json:"type" mysql:"type"`
	Payload          []byte     `json:"payload" mysql:"payload"`
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
	Payload          []byte     `json:"payload" mysql:"payload"`
	ChargeCreatedAt  time.Time  `json:"charge_created_at" mysql:"charge_created_at"`
	ChargeDeletedAt  *time.Time `json:"charge_deleted_at" mysql:"charge_deleted_at"`
}

type SQLClient struct {
	sqlClient MySQLClient
}

func NewSQLClient(client MySQLClient) *SQLClient {
	return &SQLClient{
		sqlClient: client,
	}
}

func (sql *SQLClient) CreateCharge(ctx context.Context, c billing.Charge) error {
	rawPayload, err := json.Marshal(c.Payload)
	if err != nil {
		return err
	}

	res, err := sql.sqlClient.Exec(ctx, "INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)", c.ID, string(c.Type), c.LCOrganizationID, rawPayload, time.Now())
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
		return nil, ErrChargeNotFound
	}
	var charges []*SQLCharge
	if err := res.CastTo(&charges); err != nil {
		return nil, fmt.Errorf("couldn't cast result to sql charge: %w", err)
	}
	if charges[0] == nil {
		return nil, ErrChargeNotFound
	}
	return ToBillingCharge(charges[0]), nil
}

func (sql *SQLClient) UpdateChargePayload(ctx context.Context, id string, payload livechat.BaseCharge) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := sql.sqlClient.Exec(ctx, "UPDATE charges SET payload = ? WHERE id = ? AND deleted_at IS NULL", id, rawPayload)
	if err != nil {
		return fmt.Errorf("couldn't update charge: %w", err)
	}
	if res.RowsAffected == 0 {
		return ErrChargeNotFound
	}

	return nil
}

func (sql *SQLClient) DeleteCharge(ctx context.Context, id string) error {
	res, err := sql.sqlClient.Exec(ctx, "UPDATE charges SET deleted_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		return fmt.Errorf("couldn't delete charge: %w", err)
	}
	if res.RowsAffected == 0 {
		return ErrChargeNotFound
	}

	return nil
}

func (sql *SQLClient) CreateSubscription(ctx context.Context, subscription billing.Subscription) error {
	res, err := sql.sqlClient.Exec(ctx, "INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at) VALUES (?, ?, ?, ?, ?)", subscription.ID, subscription.LCOrganizationID, subscription.PlanName, subscription.Charge.ID, time.Now())
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

func (sql *SQLClient) DeleteSubscriptionByChargeID(ctx context.Context, id string) error {
	res, err := sql.sqlClient.Exec(ctx, "UPDATE subscriptions SET deleted_at = ? WHERE charge_id = $1", time.Now(), id)
	if err != nil {
		return fmt.Errorf("couldn't delete subsctiption: %w", err)
	}
	if res.RowsAffected == 0 {
		return ErrSubscriptionNotFound
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
		Payload:          r.Payload,
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
		Payload:          c.Payload,
		CreatedAt:        c.CreatedAt,
		CanceledAt:       canceledAt,
	}
}
