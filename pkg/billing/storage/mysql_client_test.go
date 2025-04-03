package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	lcMySQL "github.com/livechat/go-mysql"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/livechat-integrations/go-billing-sdk/pkg/billing"
)

var now, _ = time.Parse("2006-01-02 15:04:05", "2025-04-02 15:04:05")
var dm = new(mysqlMock)
var cm = new(clockMock)

var mysqlClient = SQLClient{
	sqlClient: dm,
	clock:     cm,
}

var assertExpectations = func(t *testing.T) {
	mock.AssertExpectationsForObjects(t, dm, cm)
	dm.Calls = nil
	cm.Calls = nil

	dm.ExpectedCalls = nil
	cm.ExpectedCalls = nil
}

type clockMock struct {
	mock.Mock
}

func (c *clockMock) Now() time.Time {
	return c.Called().Get(0).(time.Time)
}

func (c *clockMock) After(d time.Duration) <-chan time.Time {
	//TODO implement me
	panic("implement me")
}

type mysqlMock struct {
	mock.Mock
}

func (m *mysqlMock) Exec(ctx context.Context, query string, args ...interface{}) (*lcMySQL.Meta, error) {
	tArgs := m.Called(ctx, query, args)
	if tArgs.Get(0) == nil {
		return nil, tArgs.Error(1)
	}

	return tArgs.Get(0).(*lcMySQL.Meta), tArgs.Error(1)
}

func (m *mysqlMock) Query(ctx context.Context, query string, args ...interface{}) (*lcMySQL.Results, error) {
	tArgs := m.Called(ctx, query, args)
	if tArgs.Get(0) == nil {
		return nil, tArgs.Error(1)
	}

	return tArgs.Get(0).(*lcMySQL.Results), tArgs.Error(1)
}

func (m *mysqlMock) SamplesChan() chan *lcMySQL.QueryStats {
	//TODO implement me
	panic("implement me")
}

func TestNewSQLClient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NotNil(t, NewSQLClient(&mysqlMock{}, &clockMock{}))
	})
}

func TestNewSQLClient_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		someDate, _ := time.Parse("2006-01-02", "2025-04-01")

		bch := map[string]interface{}{
			"lorem": "ipsum",
		}

		jsonPayload, _ := json.Marshal(bch)

		charge := billing.Charge{
			ID:               "id1",
			LCOrganizationID: "lcOrganizationID",
			Type:             billing.ChargeTypeRecurring,
			Payload:          jsonPayload,
			CreatedAt:        someDate,
			CanceledAt:       nil,
		}

		res := lcMySQL.Meta{
			LastIntertID: 1,
			RowsAffected: 1,
			QueryTime:    1,
		}
		rawPayload, _ := json.Marshal(charge.Payload)

		ctx := context.Background()
		cm.On("Now").Return(now).Once()
		dm.On("Exec", ctx, "INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)", []interface{}{charge.ID, string(charge.Type), charge.LCOrganizationID, rawPayload, now}).Return(&res, nil).Once()

		err := mysqlClient.CreateCharge(context.Background(), billing.Charge{
			ID:               charge.ID,
			Type:             charge.Type,
			Payload:          charge.Payload,
			LCOrganizationID: charge.LCOrganizationID,
		})
		assert.NoError(t, err)

		assertExpectations(t)
	})

	t.Run("no rows affected", func(t *testing.T) {
		someDate, _ := time.Parse("2006-01-02", "2025-04-01")

		bch := map[string]interface{}{
			"lorem": "ipsum",
		}

		jsonPayload, _ := json.Marshal(bch)

		charge := billing.Charge{
			ID:               "id1",
			LCOrganizationID: "lcOrganizationID",
			Type:             billing.ChargeTypeRecurring,
			Payload:          jsonPayload,
			CreatedAt:        someDate,
			CanceledAt:       nil,
		}

		res := lcMySQL.Meta{
			LastIntertID: 1,
			RowsAffected: 0,
			QueryTime:    1,
		}
		rawPayload, _ := json.Marshal(charge.Payload)

		ctx := context.Background()
		cm.On("Now").Return(now).Once()
		dm.On("Exec", ctx, "INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)", []interface{}{charge.ID, string(charge.Type), charge.LCOrganizationID, rawPayload, now}).Return(&res, nil).Once()

		err := mysqlClient.CreateCharge(context.Background(), billing.Charge{
			ID:               charge.ID,
			Type:             charge.Type,
			Payload:          charge.Payload,
			LCOrganizationID: charge.LCOrganizationID,
		})
		assert.Equal(t, "couldn't add new charge", err.Error())

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		someDate, _ := time.Parse("2006-01-02", "2025-04-01")

		bch := map[string]interface{}{
			"lorem": "ipsum",
		}

		jsonPayload, _ := json.Marshal(bch)

		charge := billing.Charge{
			ID:               "id1",
			LCOrganizationID: "lcOrganizationID",
			Type:             billing.ChargeTypeRecurring,
			Payload:          jsonPayload,
			CreatedAt:        someDate,
			CanceledAt:       nil,
		}

		rawPayload, _ := json.Marshal(charge.Payload)

		ctx := context.Background()
		cm.On("Now").Return(now).Once()
		dm.On("Exec", ctx, "INSERT INTO charges(id, type, payload, lc_organization_id, created_at) VALUES (?, ?, ?, ?, ?)", []interface{}{charge.ID, string(charge.Type), charge.LCOrganizationID, rawPayload, now}).Return(nil, assert.AnError).Once()

		err := mysqlClient.CreateCharge(context.Background(), billing.Charge{
			ID:               charge.ID,
			Type:             charge.Type,
			Payload:          charge.Payload,
			LCOrganizationID: charge.LCOrganizationID,
		})
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

// TODO implement more tests
