package storage

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing/storage/postgresql/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) Exec(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, s, i)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *mockDB) Query(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error) {
	args := m.Called(ctx, s, i)
	return args.Get(0).(pgx.Rows), args.Error(1)
}

func (m *mockDB) QueryRow(ctx context.Context, s string, i ...interface{}) pgx.Row {
	args := m.Called(ctx, s, i)
	return args.Get(0).(pgx.Row)
}

type rowMock struct {
	mock.Mock
}

func (m *rowMock) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

var mdb = new(mockDB)
var q = sqlc.New(mdb)
var s = NewPostgresqlSQLC(q)
var ctx = context.Background()

func TestNewPostgresqlSQLC(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NotNil(t, NewPostgresqlSQLC(q))
	})
}

func TestPostgresqlSQLC_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		mdb.On("Exec", ctx, "-- name: CreateCharge :exec\nINSERT INTO charges(id, type, payload, lc_organization_id, created_at)\nVALUES ($1, $2, $3, $4, NOW())\n", []interface{}{"1", "recurring", emptyRawPayload, "lcOrganizationID"}).Return(pgconn.CommandTag{}, nil).Once()

		err := s.CreateCharge(context.Background(), billing.Charge{
			ID:               "1",
			Type:             billing.ChargeTypeRecurring,
			Payload:          json.RawMessage("{}"),
			LCOrganizationID: "lcOrganizationID",
		})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		mdb.On("Exec", ctx, "-- name: CreateCharge :exec\nINSERT INTO charges(id, type, payload, lc_organization_id, created_at)\nVALUES ($1, $2, $3, $4, NOW())\n", []interface{}{"1", "recurring", emptyRawPayload, "lcOrganizationID"}).Return(pgconn.CommandTag{}, assert.AnError).Once()

		err := s.CreateCharge(context.Background(), billing.Charge{
			ID:               "1",
			Type:             billing.ChargeTypeRecurring,
			Payload:          json.RawMessage("{}"),
			LCOrganizationID: "lcOrganizationID",
		})
		assert.Error(t, err)
	})
}

func TestPostgresqlSQLC_CreateSubscription(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mdb.On("Exec", ctx, "-- name: CreateSubscription :exec\nINSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at)\nVALUES ($1, $2, $3, $4, NOW())\n", []interface{}{"1", "lcOrganizationID", "planName", pgtype.Text{String: "chargeID", Valid: true}}).Return(pgconn.CommandTag{}, nil).Once()

		err := s.CreateSubscription(context.Background(), billing.Subscription{
			ID:               "1",
			LCOrganizationID: "lcOrganizationID",
			PlanName:         "planName",
			Charge: &billing.Charge{
				ID: "chargeID",
			},
		})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mdb.On("Exec", ctx, "-- name: CreateSubscription :exec\nINSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at)\nVALUES ($1, $2, $3, $4, NOW())\n", []interface{}{"1", "lcOrganizationID", "planName", pgtype.Text{String: "chargeID", Valid: true}}).Return(pgconn.CommandTag{}, assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), billing.Subscription{
			ID:               "1",
			LCOrganizationID: "lcOrganizationID",
			PlanName:         "planName",
			Charge: &billing.Charge{
				ID: "chargeID",
			},
		})
		assert.Error(t, err)
	})
}

func TestPostgresqlSQLC_GetCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rm := new(rowMock)
		row := sqlc.Charge{
			ID:               "1",
			LcOrganizationID: "lcOrganizationID",
			Type:             "recurring",
			Payload:          []byte("{}"),
		}
		rm.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			a := args.Get(0).([]interface{})
			*(a[0].(*string)) = row.ID
			*(a[1].(*string)) = row.LcOrganizationID
			*(a[2].(*string)) = row.Type
			*(a[3].(*[]byte)) = row.Payload
		}).Return(nil).Once()
		mdb.On("QueryRow", ctx, "-- name: GetChargeByID :one\nSELECT id, lc_organization_id, type, payload, created_at, deleted_at\nFROM charges\nWHERE id = $1\n", []interface{}{"1"}).
			Return(rm, nil).Once()

		c, err := s.GetCharge(context.Background(), "1")
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, billing.ChargeTypeRecurring, c.Type)
		assert.Equal(t, json.RawMessage("{}"), c.Payload)
	})

	t.Run("error", func(t *testing.T) {
		mr := new(rowMock)
		mr.On("Scan", mock.Anything).Return(assert.AnError).Once()
		mdb.On("QueryRow", ctx, "-- name: GetChargeByID :one\nSELECT id, lc_organization_id, type, payload, created_at, deleted_at\nFROM charges\nWHERE id = $1\n", []interface{}{"1"}).Return(mr, assert.AnError).Once()

		_, err := s.GetCharge(context.Background(), "1")
		assert.Error(t, err)
	})
}

func TestPostgresqlSQLC_GetChargeByOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rm := new(rowMock)
		row := sqlc.Charge{
			ID:               "1",
			LcOrganizationID: "lcOrganizationID",
			Type:             "recurring",
			Payload:          []byte("{}"),
		}
		rm.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			a := args.Get(0).([]interface{})
			*(a[0].(*string)) = row.ID
			*(a[1].(*string)) = row.LcOrganizationID
			*(a[2].(*string)) = row.Type
			*(a[3].(*[]byte)) = row.Payload
		}).Return(nil).Once()
		mdb.On("QueryRow", ctx, "-- name: GetChargeByOrganizationID :one\nSELECT id, lc_organization_id, type, payload, created_at, deleted_at\nFROM charges\nWHERE lc_organization_id = $1\n", []interface{}{"lcOrganizationID"}).
			Return(rm, nil).Once()

		c, err := s.GetChargeByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, billing.ChargeTypeRecurring, c.Type)
		assert.Equal(t, json.RawMessage("{}"), c.Payload)
	})

	t.Run("error", func(t *testing.T) {
		mr := new(rowMock)
		mr.On("Scan", mock.Anything).Return(assert.AnError).Once()
		mdb.On("QueryRow", ctx, "-- name: GetChargeByOrganizationID :one\nSELECT id, lc_organization_id, type, payload, created_at, deleted_at\nFROM charges\nWHERE lc_organization_id = $1\n", []interface{}{"lcOrganizationID"}).Return(mr, assert.AnError).Once()

		_, err := s.GetChargeByOrganizationID(context.Background(), "lcOrganizationID")
		assert.Error(t, err)
	})
}

func TestPostgresqlSQLC_GetSubscriptionByOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rm := new(rowMock)
		row := sqlc.GetSubscriptionByOrganizationIDRow{
			ID:               "1",
			LcOrganizationID: "lcOrganizationID",
			PlanName:         "planName",
			ChargeID:         pgtype.Text{String: "chargeID", Valid: true},
		}
		rm.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			a := args.Get(0).([]interface{})
			*(a[0].(*string)) = row.ID
			*(a[1].(*string)) = row.LcOrganizationID
			*(a[2].(*string)) = row.PlanName
			*(a[3].(*pgtype.Text)) = row.ChargeID
		}).Return(nil).Once()
		mdb.On("QueryRow", ctx, "-- name: GetSubscriptionByOrganizationID :one\nSELECT s.id, s.lc_organization_id, plan_name, charge_id, s.created_at, s.deleted_at, c.id, c.lc_organization_id, type, payload, c.created_at, c.deleted_at\nFROM subscriptions s\nLEFT JOIN charges c on s.charge_id = c.id\nWHERE s.lc_organization_id = $1\n", []interface{}{"lcOrganizationID"}).
			Return(rm, nil).Once()

		c, err := s.GetSubscriptionByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, "planName", c.PlanName)
		assert.Equal(t, "chargeID", c.Charge.ID)
	})

	t.Run("error", func(t *testing.T) {
		mr := new(rowMock)
		mr.On("Scan", mock.Anything).Return(assert.AnError).Once()
		mdb.On("QueryRow", ctx, "-- name: GetSubscriptionByOrganizationID :one\nSELECT s.id, s.lc_organization_id, plan_name, charge_id, s.created_at, s.deleted_at, c.id, c.lc_organization_id, type, payload, c.created_at, c.deleted_at\nFROM subscriptions s\nLEFT JOIN charges c on s.charge_id = c.id\nWHERE s.lc_organization_id = $1\n", []interface{}{"lcOrganizationID"}).Return(mr, assert.AnError).Once()

		_, err := s.GetSubscriptionByOrganizationID(context.Background(), "lcOrganizationID")
		assert.Error(t, err)
	})
}

func TestPostgresqlSQLC_UpdateChargePayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(billing.BaseCharge{})
		mdb.On("Exec", ctx, "-- name: UpdateCharge :exec\nUPDATE charges\nSET payload = $2\nWHERE id = $1\n", []interface{}{"1", emptyRawPayload}).Return(pgconn.CommandTag{}, nil).Once()

		err := s.UpdateChargePayload(context.Background(), "1", billing.BaseCharge{})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(billing.BaseCharge{})
		mdb.On("Exec", ctx, "-- name: UpdateCharge :exec\nUPDATE charges\nSET payload = $2\nWHERE id = $1\n", []interface{}{"1", emptyRawPayload}).Return(pgconn.CommandTag{}, assert.AnError).Once()

		err := s.UpdateChargePayload(context.Background(), "1", billing.BaseCharge{})
		assert.Error(t, err)
	})
}
