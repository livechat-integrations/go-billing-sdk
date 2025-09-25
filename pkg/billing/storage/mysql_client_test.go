package storage

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/billing"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

func closeDB(t *testing.T, db *stdsql.DB) {
	// helper to satisfy errcheck on db.Close() in defers
	t.Helper()
	require.NoError(t, db.Close())
}

var now, _ = time.Parse("2006-01-02 15:04:05", "2025-04-02 15:04:05")

type clockMock struct {
	mock.Mock
}

func (c *clockMock) Now() time.Time {
	return c.Called().Get(0).(time.Time)
}

func (c *clockMock) After(d time.Duration) <-chan time.Time { return nil }

func TestNewSQLClient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		mock.ExpectClose()
		assert.NotNil(t, NewSQLClient(db, &clockMock{}))
		require.NoError(t, db.Close())
	})
}

func TestNewSQLClient_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bch := map[string]interface{}{"lorem": "ipsum"}
		jsonPayload, _ := json.Marshal(bch)
		charge := billing.Charge{ID: "id1", LCOrganizationID: "lcOrganizationID", Type: billing.ChargeTypeRecurring, Status: livechat.RecurrentChargeStatusPending, Payload: jsonPayload}
		cm := new(clockMock)
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		client := NewSQLClient(db, cm)
		rawPayload, _ := json.Marshal(charge.Payload)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)")).
			WithArgs(charge.ID, string(charge.Type), rawPayload, charge.LCOrganizationID, now).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		err = client.CreateCharge(context.Background(), charge)
		assert.NoError(t, err)
		require.NoError(t, db.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("no rows affected", func(t *testing.T) {
		bch := map[string]interface{}{"lorem": "ipsum"}
		jsonPayload, _ := json.Marshal(bch)
		charge := billing.Charge{ID: "id1", LCOrganizationID: "lcOrganizationID", Type: billing.ChargeTypeRecurring, Payload: jsonPayload}
		cm := new(clockMock)
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		client := NewSQLClient(db, cm)
		rawPayload, _ := json.Marshal(charge.Payload)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)")).
			WithArgs(charge.ID, string(charge.Type), rawPayload, charge.LCOrganizationID, now).
			WillReturnResult(sqlmock.NewResult(1, 0))
		mock.ExpectClose()
		err = client.CreateCharge(context.Background(), charge)
		assert.Equal(t, "couldn't add new charge", err.Error())
		require.NoError(t, db.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		bch := map[string]interface{}{"lorem": "ipsum"}
		jsonPayload, _ := json.Marshal(bch)
		charge := billing.Charge{ID: "id1", LCOrganizationID: "lcOrganizationID", Type: billing.ChargeTypeRecurring, Payload: jsonPayload}
		cm := new(clockMock)
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		client := NewSQLClient(db, cm)
		rawPayload, _ := json.Marshal(charge.Payload)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)")).
			WithArgs(charge.ID, string(charge.Type), rawPayload, charge.LCOrganizationID, now).
			WillReturnError(assert.AnError)
		mock.ExpectClose()
		err = client.CreateCharge(context.Background(), charge)
		assert.ErrorIs(t, err, assert.AnError)
		require.NoError(t, db.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})
}

func TestSQLClient_GetCharge(t *testing.T) {
	ctx := context.Background()
	id := "chg_1"

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		client := NewSQLClient(db, &clockMock{})

		rows := sqlmock.NewRows([]string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}).
			AddRow(id, "org1", string(billing.ChargeTypeRecurring), `{"foo":"bar"}`, now, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE id = ? AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnRows(rows)
		mock.ExpectClose()

		ch, err := client.GetCharge(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, id, ch.ID)
		assert.Equal(t, "org1", ch.LCOrganizationID)
		assert.Equal(t, billing.ChargeTypeRecurring, ch.Type)
		assert.Nil(t, ch.CanceledAt)
		require.NoError(t, db.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE id = ? AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnError(stdsql.ErrNoRows)
		mock.ExpectClose()
		_, err = client.GetCharge(ctx, id)
		assert.ErrorIs(t, err, billing.ErrChargeNotFound)
		require.NoError(t, db.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE id = ? AND deleted_at IS NULL")).
			WithArgs(id).
			WillReturnError(assert.AnError)
		_, err = client.GetCharge(ctx, id)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLClient_UpdateChargePayload(t *testing.T) {
	ctx := context.Background()
	id := "chg_2"
	b, _ := json.Marshal(livechat.BaseCharge{ID: id})
	payload := json.RawMessage(b)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectExec(regexp.QuoteMeta("UPDATE charges SET payload = ? WHERE id = ? AND deleted_at IS NULL")).
			WithArgs(b, id).
			WillReturnResult(sqlmock.NewResult(0, 1))
		assert.NoError(t, client.UpdateChargePayload(ctx, id, payload))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectExec(regexp.QuoteMeta("UPDATE charges SET payload = ? WHERE id = ? AND deleted_at IS NULL")).
			WithArgs(b, id).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = client.UpdateChargePayload(ctx, id, payload)
		assert.ErrorIs(t, err, billing.ErrChargeNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectExec(regexp.QuoteMeta("UPDATE charges SET payload = ? WHERE id = ? AND deleted_at IS NULL")).
			WithArgs(b, id).
			WillReturnError(assert.AnError)
		err = client.UpdateChargePayload(ctx, id, payload)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLClient_DeleteCharge(t *testing.T) {
	ctx := context.Background()
	id := "chg_3"
	cm := new(clockMock)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE charges SET deleted_at = ? WHERE id = ?")).
			WithArgs(now, id).
			WillReturnResult(sqlmock.NewResult(0, 1))
		assert.NoError(t, client.DeleteCharge(ctx, id))
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE charges SET deleted_at = ? WHERE id = ?")).
			WithArgs(now, id).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = client.DeleteCharge(ctx, id)
		assert.ErrorIs(t, err, billing.ErrChargeNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE charges SET deleted_at = ? WHERE id = ?")).
			WithArgs(now, id).
			WillReturnError(assert.AnError)
		err = client.DeleteCharge(ctx, id)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})
}

func TestSQLClient_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	sub := billing.Subscription{ID: "sub1", LCOrganizationID: "org1", PlanName: "pro", Charge: &billing.Charge{ID: "chg1"}}
	cm := new(clockMock)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at) VALUES (?, ?, ?, ?, ?)")).
			WithArgs(sub.ID, sub.LCOrganizationID, sub.PlanName, sub.Charge.ID, now).
			WillReturnResult(sqlmock.NewResult(1, 1))
		assert.NoError(t, client.CreateSubscription(ctx, sub))
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("no rows", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at) VALUES (?, ?, ?, ?, ?)")).
			WithArgs(sub.ID, sub.LCOrganizationID, sub.PlanName, sub.Charge.ID, now).
			WillReturnResult(sqlmock.NewResult(1, 0))
		err = client.CreateSubscription(ctx, sub)
		assert.EqualError(t, err, "couldn't add new subscription")
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("db err", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO subscriptions(id, lc_organization_id, plan_name, charge_id, created_at) VALUES (?, ?, ?, ?, ?)")).
			WithArgs(sub.ID, sub.LCOrganizationID, sub.PlanName, sub.Charge.ID, now).
			WillReturnError(assert.AnError)
		err = client.CreateSubscription(ctx, sub)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})
}

func TestSQLClient_GetSubscriptionsByOrganizationID(t *testing.T) {
	ctx := context.Background()
	lcID := "org1"

	t.Run("success with and without charge", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})

		cols := []string{"id", "lc_organization_id", "plan_name", "charge_id", "created_at", "deleted_at", "type", "payload", "charge_created_at", "charge_deleted_at"}
		rows := sqlmock.NewRows(cols).
			AddRow("sub1", lcID, "pro", "chg1", now, nil, string(billing.ChargeTypeRecurring), `{"a":1}`, now, nil).
			AddRow("sub2", lcID, "free", "", now, nil, "", "", time.Time{}, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT s.id, s.lc_organization_id, s.plan_name, s.charge_id, s.created_at, s.deleted_at, c.type, c.payload, c.created_at AS charge_created_at, c.deleted_at AS charge_deleted_at FROM active_subscriptions s LEFT JOIN charges c ON s.charge_id = c.id AND s.lc_organization_id = c.lc_organization_id WHERE s.lc_organization_id = ?")).
			WithArgs(lcID).
			WillReturnRows(rows)

		subs, err := client.GetSubscriptionsByOrganizationID(ctx, lcID)
		assert.NoError(t, err)
		assert.Len(t, subs, 2)
		assert.Equal(t, "sub1", subs[0].ID)
		assert.NotNil(t, subs[0].Charge)
		assert.Equal(t, "sub2", subs[1].ID)
		assert.Nil(t, subs[1].Charge)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty list", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		cols := []string{"id", "lc_organization_id", "plan_name", "charge_id", "created_at", "deleted_at", "type", "payload", "charge_created_at", "charge_deleted_at"}
		rows := sqlmock.NewRows(cols)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT s.id, s.lc_organization_id, s.plan_name, s.charge_id, s.created_at, s.deleted_at, c.type, c.payload, c.created_at AS charge_created_at, c.deleted_at AS charge_deleted_at FROM active_subscriptions s LEFT JOIN charges c ON s.charge_id = c.id AND s.lc_organization_id = c.lc_organization_id WHERE s.lc_organization_id = ?")).
			WithArgs(lcID).
			WillReturnRows(rows)
		subs, err := client.GetSubscriptionsByOrganizationID(ctx, lcID)
		assert.NoError(t, err)
		assert.Empty(t, subs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT s.id, s.lc_organization_id, s.plan_name, s.charge_id, s.created_at, s.deleted_at, c.type, c.payload, c.created_at AS charge_created_at, c.deleted_at AS charge_deleted_at FROM active_subscriptions s LEFT JOIN charges c ON s.charge_id = c.id AND s.lc_organization_id = c.lc_organization_id WHERE s.lc_organization_id = ?")).
			WithArgs(lcID).
			WillReturnError(assert.AnError)
		_, err = client.GetSubscriptionsByOrganizationID(ctx, lcID)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLClient_DeleteSubscriptionByChargeID(t *testing.T) {
	ctx := context.Background()
	lcID := "org1"
	chgID := "chg1"
	cm := new(clockMock)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET deleted_at = ? WHERE charge_id = ? AND lc_organization_id = ?")).
			WithArgs(now, chgID, lcID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		assert.NoError(t, client.DeleteSubscriptionByChargeID(ctx, lcID, chgID))
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET deleted_at = ? WHERE charge_id = ? AND lc_organization_id = ?")).
			WithArgs(now, chgID, lcID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = client.DeleteSubscriptionByChargeID(ctx, lcID, chgID)
		assert.ErrorIs(t, err, billing.ErrSubscriptionNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET deleted_at = ? WHERE charge_id = ? AND lc_organization_id = ?")).
			WithArgs(now, chgID, lcID).
			WillReturnError(assert.AnError)
		err = client.DeleteSubscriptionByChargeID(ctx, lcID, chgID)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})
}

func TestSQLClient_GetChargesByOrganizationID(t *testing.T) {
	ctx := context.Background()
	lcID := "org1"

	t.Run("success many", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		cols := []string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}
		rows := sqlmock.NewRows(cols).
			AddRow("chg1", lcID, string(billing.ChargeTypeRecurring), `{"x":1}`, now, nil).
			AddRow("chg2", lcID, string(billing.ChargeTypeRecurring), `{"y":2}`, now, &now)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE lc_organization_id = ?")).
			WithArgs(lcID).
			WillReturnRows(rows)
		chs, err := client.GetChargesByOrganizationID(ctx, lcID)
		assert.NoError(t, err)
		assert.Len(t, chs, 2)
		assert.NotNil(t, chs[1].CanceledAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		cols := []string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}
		rows := sqlmock.NewRows(cols)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE lc_organization_id = ?")).
			WithArgs(lcID).
			WillReturnRows(rows)
		chs, err := client.GetChargesByOrganizationID(ctx, lcID)
		assert.NoError(t, err)
		assert.Empty(t, chs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE lc_organization_id = ?")).
			WithArgs(lcID).
			WillReturnError(assert.AnError)
		_, err = client.GetChargesByOrganizationID(ctx, lcID)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLClient_CreateEvent(t *testing.T) {
	ctx := context.Background()
	cm := new(clockMock)
	baseEvt := events.Event{ID: "evt1", LCOrganizationID: "org1", Type: events.EventTypeInfo, Action: events.EventActionCreateCharge}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.On("Now").Return(now).Once()
		evt := baseEvt
		evt.SetPayload(map[string]any{"ok": true})
		rawPayload, _ := json.Marshal(evt.Payload)
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO billing_events(id, lc_organization_id, type, action, payload, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)")).
			WithArgs(evt.ID, evt.LCOrganizationID, string(evt.Type), string(evt.Action), rawPayload, evt.Error, now).
			WillReturnResult(sqlmock.NewResult(1, 1))
		assert.NoError(t, client.CreateEvent(ctx, evt))
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("no rows", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		evt := baseEvt
		evt.SetPayload(map[string]any{"ok": true})
		rawPayload, _ := json.Marshal(evt.Payload)
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO billing_events(id, lc_organization_id, type, action, payload, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)")).
			WithArgs(evt.ID, evt.LCOrganizationID, string(evt.Type), string(evt.Action), rawPayload, evt.Error, now).
			WillReturnResult(sqlmock.NewResult(1, 0))
		err = client.CreateEvent(ctx, evt)
		assert.EqualError(t, err, "couldn't add new billing event")
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("db err", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		evt := baseEvt
		evt.SetPayload(map[string]any{"ok": true})
		rawPayload, _ := json.Marshal(evt.Payload)
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO billing_events(id, lc_organization_id, type, action, payload, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)")).
			WithArgs(evt.ID, evt.LCOrganizationID, string(evt.Type), string(evt.Action), rawPayload, evt.Error, now).
			WillReturnError(assert.AnError)
		err = client.CreateEvent(ctx, evt)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})
}

func TestSQLClient_DeleteSubscription(t *testing.T) {
	ctx := context.Background()
	lcID := "org1"
	subID := "sub_1"
	cm := new(clockMock)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET deleted_at = ? WHERE id = ? AND lc_organization_id = ?")).
			WithArgs(now, subID, lcID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		assert.NoError(t, client.DeleteSubscription(ctx, lcID, subID))
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET deleted_at = ? WHERE id = ? AND lc_organization_id = ?")).
			WithArgs(now, subID, lcID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = client.DeleteSubscription(ctx, lcID, subID)
		assert.ErrorIs(t, err, billing.ErrSubscriptionNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, cm)
		cm.ExpectedCalls = nil
		cm.On("Now").Return(now).Once()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE subscriptions SET deleted_at = ? WHERE id = ? AND lc_organization_id = ?")).
			WithArgs(now, subID, lcID).
			WillReturnError(assert.AnError)
		err = client.DeleteSubscription(ctx, lcID, subID)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
		cm.AssertExpectations(t)
	})
}

func TestSQLClient_GetChargesByStatuses(t *testing.T) {
	ctx := context.Background()
	statuses := []string{"active", "pending"}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		cols := []string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}
		rows := sqlmock.NewRows(cols).
			AddRow("chg1", "org1", string(billing.ChargeTypeRecurring), `{"status":"active"}`, now, nil)
		// The IN clause will expand to (?,?)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE JSON_UNQUOTE(JSON_EXTRACT(payload, '$.status')) IN (?, ?) AND deleted_at IS NULL")).
			WithArgs(statuses[0], statuses[1]).
			WillReturnRows(rows)
		res, err := client.GetChargesByStatuses(ctx, statuses)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "chg1", res[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE JSON_UNQUOTE(JSON_EXTRACT(payload, '$.status')) IN (?, ?) AND deleted_at IS NULL")).
			WithArgs(statuses[0], statuses[1]).
			WillReturnRows(sqlmock.NewRows([]string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}))
		res, err := client.GetChargesByStatuses(ctx, statuses)
		assert.NoError(t, err)
		assert.Empty(t, res)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()
		client := NewSQLClient(db, &clockMock{})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges WHERE JSON_UNQUOTE(JSON_EXTRACT(payload, '$.status')) IN (?, ?) AND deleted_at IS NULL")).
			WithArgs(statuses[0], statuses[1]).
			WillReturnError(assert.AnError)
		_, err = client.GetChargesByStatuses(ctx, statuses)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
