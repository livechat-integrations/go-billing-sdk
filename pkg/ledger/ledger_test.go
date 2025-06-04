package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

var am = new(apiMock)
var sm = new(storageMock)
var xm = new(xIdMock)
var em = new(eventMock)

var s = Service{
	eventService: em,
	idProvider:   xm,
	billingAPI:   am,
	storage:      sm,
	returnURL:    "returnURL",
	masterOrgID:  "masterOrgID",
}
var ctx = context.Background()

var assertExpectations = func(t *testing.T) {
	mock.AssertExpectationsForObjects(t, em, xm, am, sm, lm)
	em.Calls = nil
	xm.Calls = nil
	am.Calls = nil
	sm.Calls = nil
	lm.Calls = nil

	em.ExpectedCalls = nil
	xm.ExpectedCalls = nil
	am.ExpectedCalls = nil
	sm.ExpectedCalls = nil
	lm.ExpectedCalls = nil
}

type eventMock struct {
	mock.Mock
}

func (m *eventMock) CreateEvent(ctx context.Context, e events.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *eventMock) ToError(ctx context.Context, params events.ToErrorParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *eventMock) ToEvent(ctx context.Context, organizationID string, action events.EventAction, eventType events.EventType, payload any) events.Event {
	args := m.Called(ctx, organizationID, action, eventType, payload)
	return args.Get(0).(events.Event)
}

type xIdMock struct {
	mock.Mock
}

func (x *xIdMock) GenerateId() string {
	args := x.Called()
	return args.Get(0).(string)
}

type apiMock struct {
	mock.Mock
}

func (m *apiMock) GetDirectCharge(ctx context.Context, id string) (*livechat.DirectCharge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*livechat.DirectCharge), args.Error(1)
}

func (m *apiMock) GetRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

func (m *apiMock) CreateDirectCharge(ctx context.Context, params livechat.CreateDirectChargeParams) (*livechat.DirectCharge, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.DirectCharge), args.Error(1)
}

func (m *apiMock) CreateRecurrentCharge(ctx context.Context, params livechat.CreateRecurrentChargeParams) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

func (m *apiMock) CancelRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

func (m *apiMock) GetRecurrentChargeV3(ctx context.Context, id string) (*livechat.RecurrentCharge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) CreateRecurrentChargeV3(ctx context.Context, params livechat.CreateRecurrentChargeParams) (*livechat.RecurrentCharge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) ActivateRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

func (m *apiMock) ActivateDirectCharge(ctx context.Context, id string) (*livechat.DirectCharge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.DirectCharge), args.Error(1)
}

type storageMock struct {
	mock.Mock
}

func (m *storageMock) GetDirectTopUpsWithoutOperations(ctx context.Context) ([]TopUp, error) {
	args := m.Called(ctx)
	return args.Get(0).([]TopUp), args.Error(1)
}

func (m *storageMock) GetRecurrentTopUpsWhereStatusNotIn(ctx context.Context, statuses []TopUpStatus) ([]TopUp, error) {
	args := m.Called(ctx, statuses)
	return args.Get(0).([]TopUp), args.Error(1)
}

func (m *storageMock) GetTopUpsByTypeWhereStatusNotIn(ctx context.Context, params GetTopUpsByTypeWhereStatusNotInParams) ([]TopUp, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]TopUp), args.Error(1)
}

func (m *storageMock) GetTopUpByIDAndOrganizationID(ctx context.Context, organizationID string, id string) (*TopUp, error) {
	args := m.Called(ctx, organizationID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func (m *storageMock) CreateLedgerOperation(ctx context.Context, c Operation) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *storageMock) GetLedgerOperations(ctx context.Context, organizationID string) ([]Operation, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Operation), args.Error(1)
}

func (m *storageMock) CreateEvent(ctx context.Context, event events.Event) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) UpsertTopUp(ctx context.Context, topUp TopUp) (*TopUp, error) {
	args := m.Called(ctx, topUp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func (m *storageMock) GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error) {
	args := m.Called(ctx, organizationID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TopUp), args.Error(1)
}

func (m *storageMock) GetTopUpByIDAndType(ctx context.Context, params GetTopUpByIDAndTypeParams) (*TopUp, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func (m *storageMock) GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]TopUp, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).(float32), args.Error(1)
}

func (m *storageMock) UpdateTopUpStatus(ctx context.Context, params UpdateTopUpStatusParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	t.Run("NewService", func(t *testing.T) {
		newService := NewService(nil, nil, nil, "labs", func(ctx context.Context) (string, error) { return "", nil }, &storageMock{}, "returnURL", "masterOrgID")

		assert.NotNil(t, newService)
		assertExpectations(t)
	})
}

func TestService_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"

		xm.On("GenerateId").Return(xid, nil)
		operation := Operation{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           -amount,
		}

		sm.On("CreateLedgerOperation", ctx, operation).Return(nil).Once()
		sc, _ := json.Marshal(operation)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateOperation,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		params := CreateChargeParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateOperation, events.EventTypeInfo, operation).Return(levent).Once()

		id, err := s.CreateCharge(context.Background(), params)

		assert.Equal(t, xid, id)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		params := CreateChargeParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
		}

		xm.On("GenerateId").Return(xid, nil)
		operation := Operation{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           -amount,
		}

		sm.On("CreateLedgerOperation", ctx, operation).Return(assert.AnError).Once()
		sc, _ := json.Marshal(params)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateOperation,
			Payload:          sc,
			Error:            "failed to create charge in database: assert.AnError general error for testing",
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateOperation, events.EventTypeInfo, operation).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create ledger operation in database: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		id, err := s.CreateCharge(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_TopUp(t *testing.T) {
	t.Run("success direct", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		payload := map[string]interface{}{"some": "payload"}
		jp, _ := json.Marshal(payload)

		operation := Operation{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           amount,
			Payload:          jp,
		}

		topUp := TopUp{
			ID:               xid,
			Amount:           amount,
			LCOrganizationID: lcoid,
			LCCharge:         jp,
			Type:             TopUpTypeDirect,
			Status:           TopUpStatusSuccess,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()
		sm.On("CreateLedgerOperation", ctx, operation).Return(nil).Once()
		sc, _ := json.Marshal(operation)
		opEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateOperation,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), opEvent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateOperation, events.EventTypeInfo, operation).Return(opEvent).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("CreateEvent", context.Background(), topEvent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, xid, id)
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("direct wrong status", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		payload := map[string]interface{}{"some": "payload"}
		jp, _ := json.Marshal(payload)

		topUp := TopUp{
			ID:               xid,
			Amount:           amount,
			LCOrganizationID: lcoid,
			LCCharge:         jp,
			Type:             TopUpTypeDirect,
			Status:           TopUpStatusProcessing,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   fmt.Errorf("top up has wrong status: processing"),
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
	t.Run("success recurrent", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		someDate, _ := time.Parse("2006-01-02", "1999-01-01")
		payload := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     xid,
				Status: "active",
			},
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate,
		}
		jp, _ := json.Marshal(payload)

		operation := Operation{
			ID:               "2341-915148800000000",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Payload:          jp,
		}

		topUp := TopUp{
			ID:                xid,
			Amount:            amount,
			LCOrganizationID:  lcoid,
			LCCharge:          jp,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusActive,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()
		sm.On("CreateLedgerOperation", ctx, operation).Return(nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(operation)
		opEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateOperation,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), opEvent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateOperation, events.EventTypeInfo, operation).Return(opEvent).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("CreateEvent", context.Background(), topEvent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "2341-915148800000000", id)
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("recurrent upsert error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		someDate, _ := time.Parse("2006-01-02", "1999-01-01")
		payload := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     xid,
				Status: "active",
			},
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate,
		}
		jp, _ := json.Marshal(payload)
		topUp := TopUp{
			ID:                xid,
			Amount:            amount,
			LCOrganizationID:  lcoid,
			LCCharge:          jp,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusActive,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(nil, assert.AnError).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
	t.Run("recurrent upsert returns nothing", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		someDate, _ := time.Parse("2006-01-02", "1999-01-01")
		payload := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     xid,
				Status: "active",
			},
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate,
		}
		jp, _ := json.Marshal(payload)
		topUp := TopUp{
			ID:                xid,
			Amount:            amount,
			LCOrganizationID:  lcoid,
			LCCharge:          jp,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusActive,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(nil, nil).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   fmt.Errorf("upsert top up error"),
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
	t.Run("recurrent no LC charge dates error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		payload := map[string]interface{}{"some": "payload"}
		jp, _ := json.Marshal(payload)
		someDate, _ := time.Parse("2006-01-02", "1999-01-01")

		topUp := TopUp{
			ID:                xid,
			Amount:            amount,
			LCOrganizationID:  lcoid,
			LCCharge:          jp,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusActive,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   fmt.Errorf("no charge at current time"),
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
	t.Run("recurrent wrong status", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		payload := map[string]interface{}{"some": "payload"}
		jp, _ := json.Marshal(payload)

		topUp := TopUp{
			ID:               xid,
			Amount:           amount,
			LCOrganizationID: lcoid,
			LCCharge:         jp,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPastDue,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   fmt.Errorf("top up has wrong status: past_due"),
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("get top up error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		payload := map[string]interface{}{"some": "payload"}
		jp, _ := json.Marshal(payload)

		topUp := TopUp{
			ID:               xid,
			Amount:           amount,
			LCOrganizationID: lcoid,
			LCCharge:         jp,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPastDue,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(nil, assert.AnError).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("no top up error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		payload := map[string]interface{}{"some": "payload"}
		jp, _ := json.Marshal(payload)

		topUp := TopUp{
			ID:               xid,
			Amount:           amount,
			LCOrganizationID: lcoid,
			LCCharge:         jp,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPastDue,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(nil, nil).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   fmt.Errorf("no existing top up in database"),
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
	t.Run("storage error", func(t *testing.T) {
		amount := float32(5.23)
		someDate, _ := time.Parse("2006-01-02", "1999-01-01")
		lcoid := "lcOrganizationID"
		payload := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     xid,
				Status: "active",
			},
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate,
		}
		jp, _ := json.Marshal(payload)

		operation := Operation{
			ID:               "2341-915148800000000",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Payload:          jp,
		}

		topUp := TopUp{
			ID:                xid,
			Amount:            amount,
			LCOrganizationID:  lcoid,
			LCCharge:          jp,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusActive,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate,
		}

		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   topUp.ID,
			Type: topUp.Type,
		}).Return(&topUp, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sm.On("CreateLedgerOperation", ctx, operation).Return(assert.AnError).Once()
		sc, _ := json.Marshal(operation)
		opEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateOperation,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateOperation, events.EventTypeInfo, operation).Return(opEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: opEvent,
			Err:   fmt.Errorf("failed to create ledger operation in database: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		sct, _ := json.Marshal(topUp)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp).Return(topEvent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: topEvent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		id, err := s.TopUp(context.Background(), topUp)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_CreateTopUpRequest(t *testing.T) {
	t.Run("success recurrent", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		confUrl := "http://livechat.com/confirmation"
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:              "id",
				Name:            "name",
				Test:            false,
				Price:           int(amount * 100),
				ConfirmationURL: confUrl,
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
			ConfirmationUrl:  confUrl,
		}
		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     int(amount * 100),
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
		}
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeRecurrent,
			Config: TopUpConfig{
				Months: &months,
			},
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()

		tu, err := s.CreateTopUpRequest(ctx, params)

		assert.Equal(t, &topUp, tu)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error recurrent no months", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeRecurrent,
			Config:         TopUpConfig{},
		}

		sc, _ := json.Marshal(params)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge V3 via lc: charge config months is nil",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create recurrent charge V3 via lc: charge config months is nil")),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error recurrent api error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeRecurrent,
			Config: TopUpConfig{
				Months: &months,
			},
		}

		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     int(amount * 100),
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge V3 via lc: assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create recurrent charge V3 via lc: %w", assert.AnError)),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error recurrent no api charge returned", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeRecurrent,
			Config: TopUpConfig{
				Months: &months,
			},
		}

		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     int(amount * 100),
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge V3 via lc: ",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create recurrent charge V3 via lc: charge is nil")),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success direct", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		rc := &livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: int(amount * 100),
			},
			Quantity: 1,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
		}
		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     int(amount * 100),
			Test:      false,
		}).Return(rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()

		sc, _ := json.Marshal(topUp)
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeDirect,
			Config:         TopUpConfig{},
		}
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, &topUp, tu)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error direct api error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeDirect,
			Config:         TopUpConfig{},
		}

		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     int(amount * 100),
			Test:      false,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create direct charge via lc: assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create direct charge via lc: %w", assert.AnError)),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error direct no api charge returned", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeDirect,
			Config:         TopUpConfig{},
		}

		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     int(amount * 100),
			Test:      false,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create direct charge via lc: charge is nil",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCreateTopUp, events.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create direct charge via lc: charge is nil")),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_GetBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"

		sm.On("GetBalance", ctx, lcoid).Return(amount, nil).Once()

		balance, err := s.GetBalance(context.Background(), lcoid)

		assert.Equal(t, amount, balance)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		lcoid := "lcOrganizationID"

		sm.On("GetBalance", ctx, lcoid).Return(float32(0), assert.AnError).Once()

		balance, err := s.GetBalance(context.Background(), lcoid)

		assert.Equal(t, float32(0), balance)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_CancelTopUpRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: int(amount * 100),
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(rc, nil).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: TopUpStatusCancelled,
		}).Return(nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "success"})

		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCancelRecurrentTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("top up not found", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(nil, nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})

		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCancelRecurrentTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
	})

	t.Run("get top up error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(nil, assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCancelRecurrentTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("api error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: int(amount * 100),
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(nil, assert.AnError).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCancelRecurrentTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: int(amount * 100),
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(rc, nil).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: TopUpStatusCancelled,
		}).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCancelRecurrentTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status not found error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: int(amount * 100),
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(rc, nil).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: TopUpStatusCancelled,
		}).Return(ErrNotFound).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})

		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCancelRecurrentTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
	})
}

func TestService_ForceCancelTopUp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: status,
		}).Return(nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": status})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionForceCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           status,
		}

		err := s.ForceCancelTopUp(context.Background(), topUp)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("not found error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: status,
		}).Return(ErrNotFound).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionForceCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           status,
		}

		err := s.ForceCancelTopUp(context.Background(), topUp)

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: status,
		}).Return(assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": status})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionForceCancelCharge,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           status,
		}

		err := s.ForceCancelTopUp(context.Background(), topUp)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_SyncTopUp(t *testing.T) {
	t.Run("success direct cancelled", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "cancelled",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusCancelled,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)

		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success direct failed", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "failed",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusFailed,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)

		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success direct declined", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "declined",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusDeclined,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)

		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success direct success", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "success",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusSuccess,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)

		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success direct processed", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "processed",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusProcessing,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success direct accepted", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "accepted",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusAccepted,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success direct pending", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "other",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success recurrent declined", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "declined",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}

		jrc, _ := json.Marshal(rc)

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusDeclined,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success recurrent cancelled", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "cancelled",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}

		jrc, _ := json.Marshal(rc)

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusCancelled,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success recurrent active", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "active",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}

		jrc, _ := json.Marshal(rc)

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusActive,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success recurrent accepted", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "accepted",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}

		jrc, _ := json.Marshal(rc)

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusAccepted,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("success recurrent pending", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "other",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}

		jrc, _ := json.Marshal(rc)

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusPending,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.Nil(t, err)
		assert.Equal(t, topUp.ID, tp.ID)
		assert.Equal(t, topUp.Amount, tp.Amount)
		assert.Equal(t, topUp.Type, tp.Type)
		assert.Equal(t, topUp.Status, topUp.Status)
		assert.Equal(t, topUp.LCCharge, topUp.LCCharge)
		assert.Equal(t, topUp.CurrentToppedUpAt, topUp.CurrentToppedUpAt)
		assert.Equal(t, topUp.NextTopUpAt, topUp.NextTopUpAt)
		assert.Equal(t, topUp.ConfirmationUrl, topUp.ConfirmationUrl)

		assertExpectations(t)
	})

	t.Run("recurrent api error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusDeclined,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, tp)

		assertExpectations(t)
	})

	t.Run("direct api error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
		}

		am.On("GetDirectCharge", ctx, "id").Return(nil, assert.AnError).Once()
		sc, _ := json.Marshal(topUp)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, tp)

		assertExpectations(t)
	})

	t.Run("no direct api charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		amount := float32(5.23)
		confUrl := "http://www.google.com/confirmation"

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
		}

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to get LC API direct top up by id: id"),
		}).Return(assert.AnError).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, tp)

		assertExpectations(t)
	})

	t.Run("no recurrent api charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		amount := float32(5.23)
		confUrl := "http://www.google.com/confirmation"

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			ConfirmationUrl:  confUrl,
		}

		am.On("GetRecurrentCharge", ctx, "id").Return(nil, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to get LC API recurrent top up by id: id"),
		}).Return(assert.AnError).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, tp)

		assertExpectations(t)
	})

	t.Run("upsert error", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "failed",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jdc, _ := json.Marshal(dc)

		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusFailed,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(nil, assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		tp, err := s.SyncTopUp(context.Background(), topUp)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, tp)

		assertExpectations(t)
	})
}

func TestService_SyncOrCancelTopUpRequests(t *testing.T) {
	t.Run("success recurrent and direct active", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Times(2)

		rc1 := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "active",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}
		rc2 := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "success",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jrc1, _ := json.Marshal(rc1)
		jrc2, _ := json.Marshal(rc2)

		topUp1 := TopUp{
			ID:                "id1",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusActive,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusSuccess,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetDirectTopUpsWithoutOperations", ctx).Return([]TopUp{}, nil).Once()
		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{topUp2}, nil).Once()
		sm.On("GetRecurrentTopUpsWhereStatusNotIn", ctx, []TopUpStatus{
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		}).Return([]TopUp{topUp1}, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id1").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id2").Return(&rc2, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp2).Return(&topUp2, nil).Once()
		sc1, _ := json.Marshal(topUp1)
		sc2, _ := json.Marshal(topUp2)
		levent1 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp1).Return(levent1).Once()
		em.On("CreateEvent", orgCtx, levent1).Return(nil).Once()
		levent2 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp2).Return(levent2).Once()
		em.On("CreateEvent", orgCtx, levent2).Return(nil).Once()

		err := s.SyncOrCancelTopUpRequests(context.Background())

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success direct without operations", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Once()

		rc2 := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "success",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jrc2, _ := json.Marshal(rc2)

		operation := Operation{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Payload:          jrc2,
		}

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusSuccess,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetDirectTopUpsWithoutOperations", ctx).Return([]TopUp{topUp2}, nil).Once()
		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{}, nil).Once()
		sm.On("GetRecurrentTopUpsWhereStatusNotIn", ctx, []TopUpStatus{
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		}).Return([]TopUp{}, nil).Once()

		sm.On("GetTopUpByIDAndType", orgCtx, GetTopUpByIDAndTypeParams{
			ID:   topUp2.ID,
			Type: topUp2.Type,
		}).Return(&topUp2, nil).Once()
		sm.On("CreateLedgerOperation", orgCtx, operation).Return(nil).Once()
		sc, _ := json.Marshal(operation)
		opEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateOperation,
			Payload:          sc,
		}
		em.On("CreateEvent", orgCtx, opEvent).Return(nil).Once()
		em.On("ToEvent", orgCtx, lcoid, events.EventActionCreateOperation, events.EventTypeInfo, operation).Return(opEvent).Once()

		sct, _ := json.Marshal(topUp2)
		topEvent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionTopUp,
			Payload:          sct,
		}
		em.On("CreateEvent", orgCtx, topEvent).Return(nil).Once()
		em.On("ToEvent", orgCtx, lcoid, events.EventActionTopUp, events.EventTypeInfo, topUp2).Return(topEvent).Once()

		err := s.SyncOrCancelTopUpRequests(context.Background())

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success recurrent and direct pending", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		now := time.Now()
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Times(2)

		rc1 := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &now,
			NextChargeAt:    &someDate2,
		}
		rc2 := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jrc1, _ := json.Marshal(rc1)
		jrc2, _ := json.Marshal(rc2)

		topUp1 := TopUp{
			ID:                "id1",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusPending,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &now,
			NextTopUpAt:       &someDate2,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		sm.On("GetDirectTopUpsWithoutOperations", ctx).Return([]TopUp{}, nil).Once()
		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{topUp2}, nil).Once()
		sm.On("GetRecurrentTopUpsWhereStatusNotIn", ctx, []TopUpStatus{
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		}).Return([]TopUp{topUp1}, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id1").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id2").Return(&rc2, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp2).Return(&topUp2, nil).Once()
		sc1, _ := json.Marshal(topUp1)
		sc2, _ := json.Marshal(topUp2)
		levent1 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp1).Return(levent1).Once()
		em.On("CreateEvent", orgCtx, levent1).Return(nil).Once()
		levent2 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp2).Return(levent2).Once()
		em.On("CreateEvent", orgCtx, levent2).Return(nil).Once()

		err := s.SyncOrCancelTopUpRequests(context.Background())

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("force cancel all old pending", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-01-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-02-14 12:31:56")
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Times(3)

		rc1 := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}
		rc2 := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jrc1, _ := json.Marshal(rc1)
		jrc2, _ := json.Marshal(rc2)

		topUp1 := TopUp{
			ID:                "id1",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusPending,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		topUp11 := TopUp{
			ID:                "id11",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusCancelled,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetDirectTopUpsWithoutOperations", ctx).Return([]TopUp{}, nil).Once()
		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{topUp2}, nil).Once()
		sm.On("GetRecurrentTopUpsWhereStatusNotIn", ctx, []TopUpStatus{
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		}).Return([]TopUp{topUp1, topUp11}, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id1").Return(&rc1, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id11").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id2").Return(&rc2, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp11).Return(&topUp11, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp2).Return(&topUp2, nil).Once()
		sm.On("UpdateTopUpStatus", orgCtx, UpdateTopUpStatusParams{
			ID:     "id1",
			Status: TopUpStatusCancelled,
		}).Return(nil).Once()
		sm.On("UpdateTopUpStatus", orgCtx, UpdateTopUpStatusParams{
			ID:     "id2",
			Status: TopUpStatusCancelled,
		}).Return(nil).Once()

		sc1, _ := json.Marshal(topUp1)
		sc1_1, _ := json.Marshal(topUp11)
		sc2, _ := json.Marshal(topUp2)
		sc11, _ := json.Marshal(map[string]interface{}{"id": "id1", "status": TopUpStatusCancelled})
		sc22, _ := json.Marshal(map[string]interface{}{"id": "id2", "status": TopUpStatusCancelled})

		levent1 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp1).Return(levent1).Once()
		em.On("CreateEvent", orgCtx, levent1).Return(nil).Once()

		levent1_1 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc1_1,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp11).Return(levent1_1).Once()
		em.On("CreateEvent", orgCtx, levent1_1).Return(nil).Once()

		levent2 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp2).Return(levent2).Once()
		em.On("CreateEvent", orgCtx, levent2).Return(nil).Once()

		levent11 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionForceCancelCharge,
			Payload:          sc11,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": "id1", "status": TopUpStatusCancelled}).Return(levent11).Once()
		em.On("CreateEvent", orgCtx, levent11).Return(nil).Once()

		levent22 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionForceCancelCharge,
			Payload:          sc22,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": "id2", "status": TopUpStatusCancelled}).Return(levent22).Once()
		em.On("CreateEvent", orgCtx, levent22).Return(nil).Once()

		err := s.SyncOrCancelTopUpRequests(context.Background())

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("should not cancel old top ups", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-01-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-02-14 12:31:56")
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Times(6)

		rc1 := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays:       0,
			Months:          months,
			CurrentChargeAt: &someDate,
			NextChargeAt:    &someDate2,
		}
		rc2 := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jrc1, _ := json.Marshal(rc1)
		jrc2, _ := json.Marshal(rc2)

		topUp1 := TopUp{
			ID:                "id1",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusActive,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}
		topUp11 := TopUp{
			ID:                "id11",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusDeclined,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}
		topUp111 := TopUp{
			ID:                "id111",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusCancelled,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusSuccess,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}
		topUp22 := TopUp{
			ID:               "id22",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusDeclined,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		topUp222 := TopUp{
			ID:               "id222",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusFailed,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetDirectTopUpsWithoutOperations", ctx).Return([]TopUp{}, nil).Once()
		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{topUp2, topUp22, topUp222}, nil).Once()

		sm.On("GetRecurrentTopUpsWhereStatusNotIn", ctx, []TopUpStatus{
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		}).Return([]TopUp{topUp1, topUp11, topUp111}, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp11).Return(&topUp11, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp111).Return(&topUp111, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp2).Return(&topUp2, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp22).Return(&topUp22, nil).Once()
		sm.On("UpsertTopUp", orgCtx, topUp222).Return(&topUp222, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id1").Return(&rc1, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id11").Return(&rc1, nil).Once()
		am.On("GetRecurrentCharge", orgCtx, "id111").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id2").Return(&rc2, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id22").Return(&rc2, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id222").Return(&rc2, nil).Once()

		sc1, _ := json.Marshal(topUp1)
		levent1 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp1).Return(levent1).Once()
		em.On("CreateEvent", orgCtx, levent1).Return(nil).Once()

		sc11, _ := json.Marshal(topUp11)
		levent11 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc11,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp11).Return(levent11).Once()
		em.On("CreateEvent", orgCtx, levent11).Return(nil).Once()

		sc111, _ := json.Marshal(topUp111)
		levent111 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc111,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp111).Return(levent111).Once()
		em.On("CreateEvent", orgCtx, levent111).Return(nil).Once()

		sc2, _ := json.Marshal(topUp2)
		levent2 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp2).Return(levent2).Once()
		em.On("CreateEvent", orgCtx, levent2).Return(nil).Once()

		sc22, _ := json.Marshal(topUp22)
		levent22 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc22,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp22).Return(levent22).Once()
		em.On("CreateEvent", orgCtx, levent22).Return(nil).Once()

		sc222, _ := json.Marshal(topUp222)
		levent222 := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc222,
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp222).Return(levent222).Once()
		em.On("CreateEvent", orgCtx, levent222).Return(nil).Once()

		err := s.SyncOrCancelTopUpRequests(context.Background())

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success recurrent and error direct", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Once()

		rc2 := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "success",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		jrc2, _ := json.Marshal(rc2)

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{topUp2}, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id2").Return(nil, assert.AnError).Once()
		sc2, _ := json.Marshal(map[string]interface{}{"id": "id2"})

		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          sc2,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, topUp2).Return(levent).Once()
		em.On("ToError", orgCtx, events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.SyncOrCancelTopUpRequests(context.Background())

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("activate accepted charges", func(t *testing.T) {
		amount := float32(5.23)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		orgCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, lcoid)
		orgCtx = context.WithValue(orgCtx, LedgerEventIDCtxKey{}, xid)
		xm.On("GenerateId").Return(xid).Times(2)

		dc := livechat.DirectCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "accepted",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}
		jdc, _ := json.Marshal(dc)
		dTopUp := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusAccepted,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		rc := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             int(amount * 100),
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "frozen",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			TrialDays: 0,
			Months:    1,
		}

		jrc, _ := json.Marshal(rc)

		rTopUp := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusFrozen,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc,
		}

		sm.On("GetTopUpsByTypeWhereStatusNotIn", ctx, GetTopUpsByTypeWhereStatusNotInParams{
			Type: TopUpTypeDirect,
			Statuses: []TopUpStatus{
				TopUpStatusSuccess,
				TopUpStatusCancelled,
				TopUpStatusFailed,
				TopUpStatusDeclined,
			},
		}).Return([]TopUp{dTopUp}, nil).Once()
		am.On("GetDirectCharge", orgCtx, "id2").Return(&dc, nil).Once()
		sm.On("UpsertTopUp", orgCtx, dTopUp).Return(&dTopUp, nil).Once()

		dp, _ := json.Marshal(dTopUp)
		dEv := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          dp,
			Error:            "",
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, dTopUp).Return(dEv).Once()
		em.On("CreateEvent", orgCtx, dEv).Return(nil).Once()

		em.On("ToEvent", orgCtx, lcoid, events.EventActionActivateCharge, events.EventTypeInfo, &dTopUp).Return(dEv).Once()
		em.On("CreateEvent", orgCtx, dEv).Return(nil).Once()

		am.On("ActivateDirectCharge", orgCtx, "id2").Return(&dc, nil).Once()

		sm.On("GetRecurrentTopUpsWhereStatusNotIn", ctx, []TopUpStatus{
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		},
		).Return([]TopUp{rTopUp}, nil).Once()

		am.On("GetRecurrentCharge", orgCtx, "id1").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", orgCtx, rTopUp).Return(&rTopUp, nil).Once()

		rp, _ := json.Marshal(rTopUp)
		rEv := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncTopUp,
			Payload:          rp,
			Error:            "",
		}
		em.On("ToEvent", orgCtx, lcoid, events.EventActionSyncTopUp, events.EventTypeInfo, rTopUp).Return(rEv).Once()
		em.On("CreateEvent", orgCtx, rEv).Return(nil).Once()

		em.On("ToEvent", orgCtx, lcoid, events.EventActionActivateCharge, events.EventTypeInfo, &rTopUp).Return(rEv).Once()
		em.On("CreateEvent", orgCtx, rEv).Return(nil).Once()

		am.On("ActivateRecurrentCharge", orgCtx, "id1").Return(&rc, nil).Once()

		sm.On("GetDirectTopUpsWithoutOperations", ctx).Return([]TopUp{}, nil).Once()

		err := s.SyncOrCancelTopUpRequests(ctx)

		assert.Nil(t, err)

		assertExpectations(t)
	})
}
