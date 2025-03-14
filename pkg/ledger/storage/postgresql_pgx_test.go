package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"github.com/livechat-integrations/go-billing-sdk/pkg/ledger"
)

var dbMock, _ = pgxmock.NewConn()
var s = NewPostgresqlPGX(dbMock)
var ctx = context.Background()

func TestNewPostgresqlSQLC(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NotNil(t, NewPostgresqlPGX(dbMock))
	})
}

func TestPostgresqlSQLC_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		status := ledger.ChargeStatusPending
		chargeType := ledger.ChargeTypeRecurrent
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))

		dbMock.ExpectExec("INSERT INTO charges").
			WithArgs(id, v, string(chargeType), string(status), emptyRawPayload, lcoid).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateCharge(context.Background(), ledger.Charge{
			ID:               id,
			Amount:           amount,
			Type:             chargeType,
			Status:           status,
			LCCharge:         json.RawMessage("{}"),
			LCOrganizationID: lcoid,
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		status := ledger.ChargeStatusPending
		chargeType := ledger.ChargeTypeRecurrent
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.
			ExpectExec("INSERT INTO charges").
			WithArgs(id, v, string(chargeType), string(status), emptyRawPayload, lcoid).Times(1).
			WillReturnError(assert.AnError)

		err := s.CreateCharge(context.Background(), ledger.Charge{
			ID:               id,
			Amount:           amount,
			Type:             chargeType,
			Status:           status,
			LCCharge:         json.RawMessage("{}"),
			LCOrganizationID: lcoid,
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_UpdateChargeStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE charges SET status").
			WithArgs(string(ledger.ChargeStatusPending), "1").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1)).Times(1)

		err := s.UpdateChargeStatus(context.Background(), "1", ledger.ChargeStatusPending)
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE charges SET status").
			WithArgs(string(ledger.ChargeStatusPending), "1").Times(1).
			WillReturnError(assert.AnError)

		err := s.UpdateChargeStatus(context.Background(), "1", ledger.ChargeStatusPending)
		assert.Error(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetChargeByIdAndType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.ChargeStatusPending
		chargeType := ledger.ChargeTypeRecurrent
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, created_at, updated_at FROM charges").
			WithArgs("1", string(chargeType), string(ledger.ChargeStatusCancelled)).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(chargeType), string(status), []byte("{}"), nil, nil)).Times(1)

		c, err := s.GetChargeByIdAndType(context.Background(), "1", chargeType)
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, amount, c.Amount)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, chargeType, c.Type)
		assert.Equal(t, status, c.Status)
		assert.Equal(t, json.RawMessage("{}"), c.LCCharge)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		chargeType := ledger.ChargeTypeRecurrent
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, created_at, updated_at FROM charges").
			WithArgs("1", string(chargeType), string(ledger.ChargeStatusCancelled)).Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetChargeByIdAndType(context.Background(), "1", chargeType)
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		chargeType := ledger.ChargeTypeRecurrent
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, created_at, updated_at FROM charges").
			WithArgs("1", string(chargeType), string(ledger.ChargeStatusCancelled)).Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetChargeByIdAndType(context.Background(), "1", chargeType)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_CreateTopUp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		status := ledger.TopUpStatusPending
		topUpType := ledger.TopUpTypeRecurrent
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		dbMock.ExpectExec("INSERT INTO top_ups").
			WithArgs(id, string(status), v, string(topUpType), lcoid, emptyRawPayload, url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateTopUp(context.Background(), ledger.TopUp{
			ID:                id,
			LCOrganizationID:  lcoid,
			Status:            status,
			Amount:            amount,
			Type:              topUpType,
			ConfirmationUrl:   url,
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			LCCharge:          json.RawMessage("{}"),
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		status := ledger.TopUpStatusPending
		topUpType := ledger.TopUpTypeRecurrent
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		dbMock.ExpectExec("INSERT INTO top_ups").
			WithArgs(id, string(status), v, string(topUpType), lcoid, emptyRawPayload, url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}).Times(1).
			WillReturnError(assert.AnError)

		err := s.CreateTopUp(context.Background(), ledger.TopUp{
			ID:                id,
			LCOrganizationID:  lcoid,
			Status:            status,
			Amount:            amount,
			Type:              topUpType,
			ConfirmationUrl:   url,
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			LCCharge:          json.RawMessage("{}"),
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("SELECT b.amount::numeric FROM").
			WithArgs("lc_organization_id", string(ledger.TopUpStatusActive), string(ledger.ChargeStatusActive)).
			WillReturnRows(
				pgxmock.NewRows([]string{"amount"}).
					AddRow(v)).Times(1)

		balance, err := s.GetBalance(context.Background(), "lc_organization_id")
		assert.NoError(t, err)
		assert.Equal(t, amount, balance)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		dbMock.ExpectQuery("SELECT b.amount::numeric FROM").
			WithArgs("lc_organization_id", string(ledger.TopUpStatusActive), string(ledger.ChargeStatusActive)).Times(1).
			WillReturnError(pgx.ErrNoRows)

		balance, err := s.GetBalance(context.Background(), "lc_organization_id")
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Equal(t, float32(0), balance)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("SELECT b.amount::numeric FROM").
			WithArgs("lc_organization_id", string(ledger.TopUpStatusActive), string(ledger.ChargeStatusActive)).Times(1).
			WillReturnError(assert.AnError)

		balance, err := s.GetBalance(context.Background(), "lc_organization_id")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, float32(0), balance)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetTopUpsByOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		c, err := s.GetTopUpsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Len(t, c, 1)
		assert.Equal(t, "1", c[0].ID)
		assert.Equal(t, status, c[0].Status)
		assert.Equal(t, amount, c[0].Amount)
		assert.Equal(t, topUpType, c[0].Type)
		assert.Equal(t, "lcOrganizationID", c[0].LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), c[0].LCCharge)
		assert.Equal(t, url, c[0].ConfirmationUrl)
		assert.Equal(t, someDate, c[0].CurrentToppedUpAt)
		assert.Equal(t, someDate2, c[0].NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(pgx.ErrNoRows)

		topUps, err := s.GetTopUpsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(assert.AnError)

		topUps, err := s.GetTopUpsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_UpdateTopUpStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE top_ups SET status").
			WithArgs(string(ledger.TopUpStatusPending), "1").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1)).Times(1)

		err := s.UpdateTopUpStatus(context.Background(), "1", ledger.TopUpStatusPending)
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE top_ups SET status").
			WithArgs(string(ledger.TopUpStatusPending), "1").Times(1).
			WillReturnError(assert.AnError)

		err := s.UpdateTopUpStatus(context.Background(), "1", ledger.TopUpStatusPending)
		assert.Error(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetTopUpByIdAndType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM top_ups WHERE id").
			WithArgs("1", string(topUpType), string(ledger.TopUpStatusCancelled)).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		c, err := s.GetTopUpByIdAndType(context.Background(), "1", topUpType)
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, status, c.Status)
		assert.Equal(t, amount, c.Amount)
		assert.Equal(t, topUpType, c.Type)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), c.LCCharge)
		assert.Equal(t, url, c.ConfirmationUrl)
		assert.Equal(t, someDate, c.CurrentToppedUpAt)
		assert.Equal(t, someDate2, c.NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		topUpType := ledger.TopUpTypeDirect
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM top_ups WHERE id").
			WithArgs("1", string(topUpType), string(ledger.TopUpStatusCancelled)).Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetTopUpByIdAndType(context.Background(), "1", topUpType)
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		topUpType := ledger.TopUpTypeDirect
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM top_ups WHERE id").
			WithArgs("1", string(topUpType), string(ledger.TopUpStatusCancelled)).Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetTopUpByIdAndType(context.Background(), "1", topUpType)
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_CreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		action := ledger.EventActionUpdateChargeStatus
		eventType := ledger.EventTypeError
		dbMock.ExpectExec("INSERT INTO events").
			WithArgs(id, lcoid, string(eventType), string(action), emptyRawPayload).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateEvent(context.Background(), ledger.Event{
			ID:               id,
			LCOrganizationID: lcoid,
			Type:             eventType,
			Action:           action,
			Payload:          json.RawMessage("{}"),
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		action := ledger.EventActionUpdateChargeStatus
		eventType := ledger.EventTypeError
		dbMock.ExpectExec("INSERT INTO events").
			WithArgs(id, lcoid, string(eventType), string(action), emptyRawPayload).Times(1).
			WillReturnError(assert.AnError)

		err := s.CreateEvent(context.Background(), ledger.Event{
			ID:               id,
			LCOrganizationID: lcoid,
			Type:             eventType,
			Action:           action,
			Payload:          json.RawMessage("{}"),
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}
