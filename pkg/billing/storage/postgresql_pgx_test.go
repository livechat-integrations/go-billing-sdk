package storage

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/billing"
)

var dbMock, _ = pgxmock.NewConn()
var s = NewPostgresqlPGX(dbMock)

func TestNewPostgresqlSQLC(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NotNil(t, NewPostgresqlPGX(dbMock))
	})
}

func TestPostgresqlSQLC_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		dbMock.ExpectExec("INSERT INTO charges").
			WithArgs("1", "recurring", emptyRawPayload, "lcOrganizationID").
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateCharge(context.Background(), billing.Charge{
			ID:               "1",
			Type:             billing.ChargeTypeRecurring,
			Payload:          json.RawMessage("{}"),
			LCOrganizationID: "lcOrganizationID",
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		dbMock.
			ExpectExec("INSERT INTO charges").
			WithArgs("1", "recurring", emptyRawPayload, "lcOrganizationID").Times(1).
			WillReturnError(assert.AnError)

		err := s.CreateCharge(context.Background(), billing.Charge{
			ID:               "1",
			Type:             billing.ChargeTypeRecurring,
			Payload:          json.RawMessage("{}"),
			LCOrganizationID: "lcOrganizationID",
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_CreateSubscription(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.
			ExpectExec("INSERT INTO subscriptions").
			WithArgs("1", "lcOrganizationID", "planName", pgtype.Text{String: "chargeID", Valid: true}).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

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
		dbMock.ExpectExec("INSERT INTO subscriptions").
			WithArgs("1", "lcOrganizationID", "planName", pgtype.Text{String: "chargeID", Valid: true}).Times(1).
			WillReturnError(assert.AnError)

		err := s.CreateSubscription(context.Background(), billing.Subscription{
			ID:               "1",
			LCOrganizationID: "lcOrganizationID",
			PlanName:         "planName",
			Charge: &billing.Charge{
				ID: "chargeID",
			},
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("1").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}).
					AddRow("1", "lcOrganizationID", "recurring", []byte("{}"), nil, nil)).Times(1)

		c, err := s.GetCharge(context.Background(), "1")
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, billing.ChargeTypeRecurring, c.Type)
		assert.Equal(t, json.RawMessage("{}"), c.Payload)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("1").Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetCharge(context.Background(), "1")
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("1").Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetCharge(context.Background(), "1")
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetChargeByOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("lcOrganizationID").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}).
					AddRow("1", "lcOrganizationID", "recurring", []byte("{}"), nil, nil)).Times(1)

		c, err := s.GetChargeByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, billing.ChargeTypeRecurring, c.Type)
		assert.Equal(t, json.RawMessage("{}"), c.Payload)
	})

	t.Run("no rows", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetChargeByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetChargeByOrganizationID(context.Background(), "lcOrganizationID")
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetSubscriptionsByOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT s.id, s.lc_organization_id, plan_name, charge_id, s.created_at, s.deleted_at, c.id, c.lc_organization_id, type, payload, c.created_at, c.deleted_at FROM active_subscriptions s LEFT JOIN charges c on s.charge_id = c.id").
			WithArgs("lcOrganizationID").
			WillReturnRows(pgxmock.NewRows([]string{"id", "lc_organization_id", "plan_name", "charge_id", "created_at", "deleted_at", "id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}).
				AddRow("1", "lcOrganizationID", "planName", "chargeID", nil, nil, "chargeID", "lcOrganizationID", "recurring", []byte("{}"), nil, nil)).Times(1)

		c, err := s.GetSubscriptionsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
		assert.Equal(t, "1", c[0].ID)
		assert.Equal(t, "lcOrganizationID", c[0].LCOrganizationID)
		assert.Equal(t, "planName", c[0].PlanName)
		assert.Equal(t, "chargeID", c[0].Charge.ID)
	})

	t.Run("no rows", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT s.id, s.lc_organization_id, plan_name, charge_id, s.created_at, s.deleted_at, c.id, c.lc_organization_id, type, payload, c.created_at, c.deleted_at FROM active_subscriptions s LEFT JOIN charges c on s.charge_id = c.id").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetSubscriptionsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT s.id, s.lc_organization_id, plan_name, charge_id, s.created_at, s.deleted_at, c.id, c.lc_organization_id, type, payload, c.created_at, c.deleted_at FROM active_subscriptions s LEFT JOIN charges c on s.charge_id = c.id").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetSubscriptionsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_UpdateChargePayload(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(livechat.BaseCharge{})
		dbMock.ExpectExec("UPDATE charges SET payload").
			WithArgs("1", emptyRawPayload).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1)).Times(1)

		err := s.UpdateChargePayload(context.Background(), "1", emptyRawPayload)
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("0 rows", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(livechat.BaseCharge{})
		dbMock.ExpectExec("UPDATE charges SET payload").
			WithArgs("1", emptyRawPayload).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0)).Times(1)

		err := s.UpdateChargePayload(context.Background(), "1", emptyRawPayload)
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(livechat.BaseCharge{})
		dbMock.ExpectExec("UPDATE charges SET payload").
			WithArgs("1", emptyRawPayload).Times(1).
			WillReturnError(assert.AnError)

		err := s.UpdateChargePayload(context.Background(), "1", emptyRawPayload)
		assert.Error(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlPGX_DeleteCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE charges SET deleted_at").
			WithArgs("1").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1)).Times(1)

		err := s.DeleteCharge(context.Background(), "1")
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE charges SET deleted_at").
			WithArgs("1").Times(1).
			WillReturnError(assert.AnError)

		err := s.DeleteCharge(context.Background(), "1")
		assert.Error(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlPGX_DeleteSubscriptionByChargeID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE subscriptions SET deleted_at = now()").
			WithArgs(pgtype.Text{String: "1", Valid: true}, "lcoid").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1)).Times(1)

		err := s.DeleteSubscriptionByChargeID(context.Background(), "lcoid", "1")
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE subscriptions SET deleted_at = now()").
			WithArgs(pgtype.Text{String: "1", Valid: true}, "lcoid").Times(1).
			WillReturnError(assert.AnError)

		err := s.DeleteSubscriptionByChargeID(context.Background(), "lcoid", "1")
		assert.Error(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlPGX_GetChargesByOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("lcOrganizationID").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "lc_organization_id", "type", "payload", "created_at", "deleted_at"}).
					AddRow("1", "lcOrganizationID", "recurring", []byte("{}"), nil, nil)).Times(1)

		c, err := s.GetChargesByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
		assert.Equal(t, "1", c[0].ID)
		assert.Equal(t, "lcOrganizationID", c[0].LCOrganizationID)
		assert.Equal(t, billing.ChargeTypeRecurring, c[0].Type)
		assert.Equal(t, json.RawMessage("{}"), c[0].Payload)
	})

	t.Run("no rows", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetChargesByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT id, lc_organization_id, type, payload, created_at, deleted_at FROM charges").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetChargesByOrganizationID(context.Background(), "lcOrganizationID")
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}
