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

	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/ledger"
)

var dbMock, _ = pgxmock.NewConn()
var s = NewPostgresqlPGX(dbMock)

func TestNewPostgresqlSQLC(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NotNil(t, NewPostgresqlPGX(dbMock))
	})
}

func TestPostgresqlSQLC_CreateLedgerOperation(t *testing.T) {
	t.Run("success positive", func(t *testing.T) {
		id := "1"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		payload := map[string]interface{}{"some": "field"}
		jp, _ := json.Marshal(payload)

		dbMock.ExpectExec("INSERT INTO ledger_ledger").
			WithArgs(id, v, lcoid, jp).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateLedgerOperation(context.Background(), ledger.Operation{
			ID:               id,
			Amount:           amount,
			Payload:          jp,
			LCOrganizationID: lcoid,
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("success negative", func(t *testing.T) {
		id := "1"
		lcoid := "lcOrganizationID"
		amount := -float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		payload := map[string]interface{}{"some": "field"}
		jp, _ := json.Marshal(payload)

		dbMock.ExpectExec("INSERT INTO ledger_ledger").
			WithArgs(id, v, lcoid, jp).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateLedgerOperation(context.Background(), ledger.Operation{
			ID:               id,
			Amount:           amount,
			Payload:          jp,
			LCOrganizationID: lcoid,
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("error", func(t *testing.T) {
		id := "1"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		payload := map[string]interface{}{"some": "field"}
		jp, _ := json.Marshal(payload)

		dbMock.ExpectExec("INSERT INTO ledger_ledger").
			WithArgs(id, v, lcoid, jp).Times(1).WillReturnError(assert.AnError)

		err := s.CreateLedgerOperation(context.Background(), ledger.Operation{
			ID:               id,
			Amount:           amount,
			Payload:          jp,
			LCOrganizationID: lcoid,
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetLedgerOperations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		dbMock.ExpectQuery("GetLedgerOperationsByOrganizationID :many SELECT id, amount, lc_organization_id, payload, created_at").
			WithArgs("lcOrganizationID").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "payload", "created_at"}).
					AddRow("1", v, "lcOrganizationID", []byte("{}"), pgtype.Timestamptz{Time: someDate, Valid: true})).Times(1)

		c, err := s.GetLedgerOperations(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Len(t, c, 1)
		assert.Equal(t, "1", c[0].ID)
		assert.Equal(t, amount, c[0].Amount)
		assert.Equal(t, "lcOrganizationID", c[0].LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), c[0].Payload)
		assert.Equal(t, someDate, c[0].CreatedAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("no records", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetLedgerOperationsByOrganizationID :many SELECT id, amount, lc_organization_id, payload, created_at").
			WithArgs("lcOrganizationID").
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "payload", "created_at"})).Times(1)

		c, err := s.GetLedgerOperations(context.Background(), "lcOrganizationID")
		assert.NoError(t, err)
		assert.Len(t, c, 0)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetLedgerOperationsByOrganizationID :many SELECT id, amount, lc_organization_id, payload, created_at").
			WithArgs("lcOrganizationID").Times(1).WillReturnError(assert.AnError)

		c, err := s.GetLedgerOperations(context.Background(), "lcOrganizationID")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, c, 0)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetLedgerOperation(t *testing.T) {
	id := "1"
	lcoid := "lcOrganizationID"
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		dbMock.ExpectQuery("GetLedgerOperation :one SELECT id, amount, lc_organization_id, payload, created_at").
			WithArgs(lcoid, id).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "payload", "created_at"}).
					AddRow(id, v, lcoid, []byte("{}"), pgtype.Timestamptz{Time: someDate, Valid: true})).Times(1)

		c, err := s.GetLedgerOperation(context.Background(), ledger.GetLedgerOperationParams{
			ID:             id,
			OrganizationID: lcoid,
		})
		assert.NoError(t, err)
		assert.Equal(t, id, c.ID)
		assert.Equal(t, amount, c.Amount)
		assert.Equal(t, lcoid, c.LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), c.Payload)
		assert.Equal(t, someDate, c.CreatedAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("no records", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetLedgerOperation :one SELECT id, amount, lc_organization_id, payload, created_at").
			WithArgs(lcoid, id).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "payload", "created_at"})).Times(1)

		c, err := s.GetLedgerOperation(context.Background(), ledger.GetLedgerOperationParams{
			ID:             id,
			OrganizationID: lcoid,
		})
		assert.NoError(t, err)
		assert.Nil(t, c)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetLedgerOperation :one SELECT id, amount, lc_organization_id, payload, created_at").
			WithArgs(lcoid, id).
			Times(1).WillReturnError(assert.AnError)
		c, err := s.GetLedgerOperation(context.Background(), ledger.GetLedgerOperationParams{
			ID:             id,
			OrganizationID: lcoid,
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, c)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_UpsertTopUp(t *testing.T) {
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

		dbMock.ExpectQuery("UpsertTopUp :one INSERT INTO ledger_top_ups").
			WithArgs(id, string(status), v, string(topUpType), lcoid, emptyRawPayload, url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		ut, err := s.UpsertTopUp(context.Background(), ledger.TopUp{
			ID:                id,
			LCOrganizationID:  lcoid,
			Status:            status,
			Amount:            amount,
			Type:              topUpType,
			ConfirmationUrl:   url,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
			LCCharge:          json.RawMessage("{}"),
		})
		assert.NoError(t, err)
		assert.Equal(t, id, ut.ID)
		assert.Equal(t, status, ut.Status)
		assert.Equal(t, amount, ut.Amount)
		assert.Equal(t, lcoid, ut.LCOrganizationID)
		assert.Equal(t, url, ut.ConfirmationUrl)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error not found", func(t *testing.T) {
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

		dbMock.ExpectQuery("UpsertTopUp :one INSERT INTO ledger_top_ups").
			WithArgs(id, string(status), v, string(topUpType), lcoid, emptyRawPayload, url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}).
			Times(1).WillReturnError(pgx.ErrNoRows)

		_, err := s.UpsertTopUp(context.Background(), ledger.TopUp{
			ID:                id,
			LCOrganizationID:  lcoid,
			Status:            status,
			Amount:            amount,
			Type:              topUpType,
			ConfirmationUrl:   url,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
			LCCharge:          json.RawMessage("{}"),
		})
		assert.ErrorIs(t, err, ledger.ErrNotFound)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetOrganizationBalance :one SELECT").
			WithArgs("lc_organization_id").
			WillReturnRows(
				pgxmock.NewRows([]string{"amount"}).
					AddRow(v)).Times(1)

		balance, err := s.GetBalance(context.Background(), "lc_organization_id")
		assert.NoError(t, err)
		assert.Equal(t, amount, balance)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		dbMock.ExpectQuery("GetOrganizationBalance :one SELECT").
			WithArgs("lc_organization_id").Times(1).
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
		dbMock.ExpectQuery("GetOrganizationBalance :one SELECT").
			WithArgs("lc_organization_id").Times(1).
			WillReturnError(assert.AnError)

		balance, err := s.GetBalance(context.Background(), "lc_organization_id")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, float32(0), balance)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlPGX_GetTopUpsByIDAndOrganizationID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		id := "id"
		lcoid := "lcOrganizationID"
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeDirect
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("GetTopUpByIDAndOrganizationID :one SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs(id, lcoid).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow(id, v, lcoid, string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		topUp, err := s.GetTopUpByIDAndOrganizationID(context.Background(), lcoid, id)
		assert.NoError(t, err)
		assert.Equal(t, id, topUp.ID)
		assert.Equal(t, amount, topUp.Amount)
		assert.Equal(t, lcoid, topUp.LCOrganizationID)
		assert.Equal(t, status, topUp.Status)
		assert.Equal(t, topUpType, topUp.Type)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("no top up error", func(t *testing.T) {
		id := "id"
		lcoid := "lcOrganizationID"
		dbMock.ExpectQuery("GetTopUpByIDAndOrganizationID :one SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs(id, lcoid).Times(1).
			WillReturnError(pgx.ErrNoRows)

		topUp, err := s.GetTopUpByIDAndOrganizationID(context.Background(), lcoid, id)
		assert.NoError(t, err)
		assert.Nil(t, topUp)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
	t.Run("error", func(t *testing.T) {
		id := "id"
		lcoid := "lcOrganizationID"
		dbMock.ExpectQuery("GetTopUpByIDAndOrganizationID :one SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs(id, lcoid).Times(1).
			WillReturnError(assert.AnError)

		topUp, err := s.GetTopUpByIDAndOrganizationID(context.Background(), lcoid, id)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, topUp)

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
		dbMock.ExpectQuery("GetTopUpsByOrganizationID :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups").
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
		assert.Equal(t, someDate, *c[0].CurrentToppedUpAt)
		assert.Equal(t, someDate2, *c[0].NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE lc_organization_id").
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
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID").Times(1).
			WillReturnError(assert.AnError)

		topUps, err := s.GetTopUpsByOrganizationID(context.Background(), "lcOrganizationID")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlPGX_GetTopUpsByTypeWhereStatusNotIn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("GetTopUpsByTypeWhereStatusNotIn :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs(string(topUpType), []string{string(ledger.TopUpStatusCancelled), string(ledger.TopUpStatusFailed)}).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		topUps, err := s.GetTopUpsByTypeWhereStatusNotIn(context.Background(), ledger.GetTopUpsByTypeWhereStatusNotInParams{
			Type:     ledger.TopUpTypeDirect,
			Statuses: []ledger.TopUpStatus{ledger.TopUpStatusCancelled, ledger.TopUpStatusFailed},
		})
		assert.NoError(t, err)
		assert.Len(t, topUps, 1)
		assert.Equal(t, "1", topUps[0].ID)
		assert.Equal(t, status, topUps[0].Status)
		assert.Equal(t, amount, topUps[0].Amount)
		assert.Equal(t, topUpType, topUps[0].Type)
		assert.Equal(t, "lcOrganizationID", topUps[0].LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), topUps[0].LCCharge)
		assert.Equal(t, url, topUps[0].ConfirmationUrl)
		assert.Equal(t, someDate, *topUps[0].CurrentToppedUpAt)
		assert.Equal(t, someDate2, *topUps[0].NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		amount := float32(3.14)
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetTopUpsByTypeWhereStatusNotIn :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs(string(topUpType), []string{string(ledger.TopUpStatusCancelled), string(ledger.TopUpStatusFailed)}).Times(1).
			WillReturnError(pgx.ErrNoRows)
		topUps, err := s.GetTopUpsByTypeWhereStatusNotIn(context.Background(), ledger.GetTopUpsByTypeWhereStatusNotInParams{
			Type:     ledger.TopUpTypeDirect,
			Statuses: []ledger.TopUpStatus{ledger.TopUpStatusCancelled, ledger.TopUpStatusFailed},
		})
		assert.NoError(t, err)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetTopUpsByTypeWhereStatusNotIn :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs(string(topUpType), []string{string(ledger.TopUpStatusCancelled), string(ledger.TopUpStatusFailed)}).Times(1).
			WillReturnError(assert.AnError)
		topUps, err := s.GetTopUpsByTypeWhereStatusNotIn(context.Background(), ledger.GetTopUpsByTypeWhereStatusNotInParams{
			Type:     ledger.TopUpTypeDirect,
			Statuses: []ledger.TopUpStatus{ledger.TopUpStatusCancelled, ledger.TopUpStatusFailed},
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlPGX_GetRecurrentTopUpsWhereStatusNotIn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeRecurrent
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("GetRecurrentTopUpsWhereStatusNotIn :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs([]string{string(ledger.TopUpStatusCancelled), string(ledger.TopUpStatusFailed)}).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		topUps, err := s.GetRecurrentTopUpsWhereStatusNotIn(context.Background(), []ledger.TopUpStatus{ledger.TopUpStatusCancelled, ledger.TopUpStatusFailed})
		assert.NoError(t, err)
		assert.Len(t, topUps, 1)
		assert.Equal(t, "1", topUps[0].ID)
		assert.Equal(t, status, topUps[0].Status)
		assert.Equal(t, amount, topUps[0].Amount)
		assert.Equal(t, topUpType, topUps[0].Type)
		assert.Equal(t, "lcOrganizationID", topUps[0].LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), topUps[0].LCCharge)
		assert.Equal(t, url, topUps[0].ConfirmationUrl)
		assert.Equal(t, someDate, *topUps[0].CurrentToppedUpAt)
		assert.Equal(t, someDate2, *topUps[0].NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetRecurrentTopUpsWhereStatusNotIn :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs([]string{string(ledger.TopUpStatusCancelled), string(ledger.TopUpStatusFailed)}).Times(1).
			WillReturnError(pgx.ErrNoRows)
		topUps, err := s.GetRecurrentTopUpsWhereStatusNotIn(context.Background(), []ledger.TopUpStatus{ledger.TopUpStatusCancelled, ledger.TopUpStatusFailed})
		assert.NoError(t, err)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetRecurrentTopUpsWhereStatusNotIn :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at").
			WithArgs([]string{string(ledger.TopUpStatusCancelled), string(ledger.TopUpStatusFailed)}).Times(1).
			WillReturnError(assert.AnError)
		topUps, err := s.GetRecurrentTopUpsWhereStatusNotIn(context.Background(), []ledger.TopUpStatus{ledger.TopUpStatusCancelled, ledger.TopUpStatusFailed})
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetTopUpsByOrganizationIDAndStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("GetTopUpsByOrganizationIDAndStatus :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID", string(status)).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		c, err := s.GetTopUpsByOrganizationIDAndStatus(context.Background(), "lcOrganizationID", status)
		assert.NoError(t, err)
		assert.Len(t, c, 1)
		assert.Equal(t, "1", c[0].ID)
		assert.Equal(t, status, c[0].Status)
		assert.Equal(t, amount, c[0].Amount)
		assert.Equal(t, topUpType, c[0].Type)
		assert.Equal(t, "lcOrganizationID", c[0].LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), c[0].LCCharge)
		assert.Equal(t, url, c[0].ConfirmationUrl)
		assert.Equal(t, someDate, *c[0].CurrentToppedUpAt)
		assert.Equal(t, someDate2, *c[0].NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetTopUpsByOrganizationIDAndStatus :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID", string(status)).Times(1).
			WillReturnError(pgx.ErrNoRows)

		topUps, err := s.GetTopUpsByOrganizationIDAndStatus(context.Background(), "lcOrganizationID", status)
		assert.NoError(t, err)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		dbMock.ExpectQuery("GetTopUpsByOrganizationIDAndStatus :many SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE lc_organization_id").
			WithArgs("lcOrganizationID", string(status)).Times(1).
			WillReturnError(assert.AnError)

		topUps, err := s.GetTopUpsByOrganizationIDAndStatus(context.Background(), "lcOrganizationID", status)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, topUps, 0)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_UpdateTopUpStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE ledger_top_ups SET status").
			WithArgs(string(ledger.TopUpStatusPending), "1").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1)).Times(1)

		err := s.UpdateTopUpStatus(context.Background(), ledger.UpdateTopUpStatusParams{
			ID:     "1",
			Status: ledger.TopUpStatusPending,
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE ledger_top_ups SET status").
			WithArgs(string(ledger.TopUpStatusPending), "1").Times(1).
			WillReturnError(assert.AnError)

		err := s.UpdateTopUpStatus(context.Background(), ledger.UpdateTopUpStatusParams{
			ID:     "1",
			Status: ledger.TopUpStatusPending,
		})
		assert.Error(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		dbMock.ExpectExec("UPDATE ledger_top_ups SET status").
			WithArgs(string(ledger.TopUpStatusPending), "1").Times(1).
			WillReturnError(pgx.ErrNoRows)

		err := s.UpdateTopUpStatus(context.Background(), ledger.UpdateTopUpStatusParams{
			ID:     "1",
			Status: ledger.TopUpStatusPending,
		})
		assert.ErrorIs(t, err, ledger.ErrNotFound)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_GetTopUpByIDAndType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(3.14)
		status := ledger.TopUpStatusActive
		topUpType := ledger.TopUpTypeDirect
		v := pgtype.Numeric{}
		_ = v.Scan(fmt.Sprintf("%f", amount))
		url := "some_url"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE id").
			WithArgs("1", string(topUpType), string(ledger.TopUpStatusCancelled)).
			WillReturnRows(
				pgxmock.NewRows([]string{"id", "amount", "lc_organization_id", "type", "status", "lc_charge", "confirmation_url", "current_topped_up_at", "next_top_up_at", "created_at", "updated_at"}).
					AddRow("1", v, "lcOrganizationID", string(topUpType), string(status), []byte("{}"), url, pgtype.Timestamptz{Time: someDate, Valid: true}, pgtype.Timestamptz{Time: someDate2, Valid: true}, nil, nil)).Times(1)

		c, err := s.GetTopUpByIDAndType(context.Background(), ledger.GetTopUpByIDAndTypeParams{
			ID:   "1",
			Type: topUpType,
		})
		assert.NoError(t, err)
		assert.Equal(t, "1", c.ID)
		assert.Equal(t, status, c.Status)
		assert.Equal(t, amount, c.Amount)
		assert.Equal(t, topUpType, c.Type)
		assert.Equal(t, "lcOrganizationID", c.LCOrganizationID)
		assert.Equal(t, json.RawMessage("{}"), c.LCCharge)
		assert.Equal(t, url, c.ConfirmationUrl)
		assert.Equal(t, someDate, *c.CurrentToppedUpAt)
		assert.Equal(t, someDate2, *c.NextTopUpAt)

		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		topUpType := ledger.TopUpTypeDirect
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE id").
			WithArgs("1", string(topUpType), string(ledger.TopUpStatusCancelled)).Times(1).
			WillReturnError(pgx.ErrNoRows)

		c, err := s.GetTopUpByIDAndType(context.Background(), ledger.GetTopUpByIDAndTypeParams{
			ID:   "1",
			Type: topUpType,
		})
		assert.NoError(t, err)
		assert.Nil(t, c)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		topUpType := ledger.TopUpTypeDirect
		dbMock.ExpectQuery("SELECT id, amount, lc_organization_id, type, status, lc_charge, confirmation_url, current_topped_up_at, next_top_up_at, created_at, updated_at FROM ledger_top_ups WHERE id").
			WithArgs("1", string(topUpType), string(ledger.TopUpStatusCancelled)).Times(1).
			WillReturnError(assert.AnError)

		_, err := s.GetTopUpByIDAndType(context.Background(), ledger.GetTopUpByIDAndTypeParams{
			ID:   "1",
			Type: topUpType,
		})
		assert.ErrorIs(t, err, assert.AnError)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})
}

func TestPostgresqlSQLC_CreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		action := events.EventActionForceCancelCharge
		eventType := events.EventTypeError
		em := "lorem ipsum"
		dbMock.ExpectExec("INSERT INTO ledger_events").
			WithArgs(id, lcoid, string(eventType), string(action), emptyRawPayload, pgtype.Text{
				String: em,
				Valid:  true,
			}).
			WillReturnResult(pgxmock.NewResult("INSERT", 1)).Times(1)

		err := s.CreateEvent(context.Background(), events.Event{
			ID:               id,
			LCOrganizationID: lcoid,
			Type:             eventType,
			Action:           action,
			Payload:          json.RawMessage("{}"),
			Error:            em,
		})
		assert.NoError(t, err)
		assert.NoError(t, dbMock.ExpectationsWereMet())
	})

	t.Run("error", func(t *testing.T) {
		emptyRawPayload, _ := json.Marshal(json.RawMessage("{}"))
		id := "1"
		lcoid := "lcOrganizationID"
		action := events.EventActionForceCancelCharge
		eventType := events.EventTypeError
		em := ""
		dbMock.ExpectExec("INSERT INTO ledger_events").
			WithArgs(id, lcoid, string(eventType), string(action), emptyRawPayload, pgtype.Text{
				String: em,
				Valid:  true,
			}).Times(1).
			WillReturnError(assert.AnError)

		err := s.CreateEvent(context.Background(), events.Event{
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
