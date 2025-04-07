package common

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var sm = new(storageMock)
var xm = new(xIdMock)

var s = Service{
	idProvider:    xm,
	storage:       sm,
	eventIdCtxKey: "ledgerCTXKey",
}
var ctx = context.Background()

var assertExpectations = func(t *testing.T) {
	mock.AssertExpectationsForObjects(t, xm, sm)
	xm.Calls = nil
	sm.Calls = nil

	xm.ExpectedCalls = nil
	sm.ExpectedCalls = nil
}

type xIdMock struct {
	mock.Mock
}

func (x *xIdMock) GenerateId() string {
	args := x.Called()
	return args.Get(0).(string)
}

type storageMock struct {
	mock.Mock
}

func (m *storageMock) CreateEvent(ctx context.Context, event Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	t.Run("NewService", func(t *testing.T) {
		newService := NewService(&storageMock{}, &xIdMock{}, "ledgerCTXKey")

		assert.NotNil(t, newService)
		assertExpectations(t)
	})
}

func TestService_ToEvent(t *testing.T) {
	t.Run("success id from context", func(t *testing.T) {
		id := "id-from-context"
		localCtx := context.WithValue(context.Background(), "ledgerCTXKey", id)
		lcoid := "lcOrganizationID"
		action := EventActionSyncTopUp
		eventType := EventTypeInfo
		payload := map[string]interface{}{"lorem": "ipsum"}

		event := s.ToEvent(localCtx, lcoid, action, eventType, payload)

		assert.Equal(t, eventType, event.Type)
		assert.Equal(t, id, event.ID)
		assert.Equal(t, action, event.Action)
		assert.Equal(t, lcoid, event.LCOrganizationID)

		assertExpectations(t)
	})
	t.Run("success no id in context", func(t *testing.T) {
		id := "id-not-from-context"
		lcoid := "lcOrganizationID"
		action := EventActionSyncTopUp
		eventType := EventTypeInfo
		payload := map[string]interface{}{"lorem": "ipsum"}

		xm.On("GenerateId").Return(id, nil)
		event := s.ToEvent(context.Background(), lcoid, action, eventType, payload)

		assert.Equal(t, eventType, event.Type)
		assert.Equal(t, id, event.ID)
		assert.Equal(t, action, event.Action)
		assert.Equal(t, lcoid, event.LCOrganizationID)

		assertExpectations(t)
	})
}

func TestService_CreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		event := Event{
			ID:               "id",
			LCOrganizationID: "lcOrganizationID",
			Type:             EventTypeInfo,
			Action:           EventActionForceCancelCharge,
			Payload:          json.RawMessage{},
			CreatedAt:        time.Time{},
		}

		sm.On("CreateEvent", context.Background(), event).Return(nil).Once()
		err := s.CreateEvent(context.Background(), event)

		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		event := Event{
			ID:               "id",
			LCOrganizationID: "lcOrganizationID",
			Type:             EventTypeInfo,
			Action:           EventActionForceCancelCharge,
			Payload:          json.RawMessage{},
			CreatedAt:        time.Time{},
		}

		sm.On("CreateEvent", context.Background(), event).Return(assert.AnError).Once()
		err := s.CreateEvent(context.Background(), event)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_ToError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		event := Event{
			ID:               "id",
			LCOrganizationID: "lcOrganizationID",
			Type:             EventTypeError,
			Action:           EventActionForceCancelCharge,
			Payload:          json.RawMessage{},
			Error:            "assert.AnError general error for testing",
			CreatedAt:        time.Time{},
		}

		sm.On("CreateEvent", context.Background(), event).Return(nil).Once()
		err := s.ToError(context.Background(), ToErrorParams{
			Event: event,
			Err:   assert.AnError,
		})

		assert.Equal(t, "id: assert.AnError general error for testing", err.Error())

		assertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		event := Event{
			ID:               "id",
			LCOrganizationID: "lcOrganizationID",
			Type:             EventTypeError,
			Action:           EventActionForceCancelCharge,
			Payload:          json.RawMessage{},
			Error:            "some error",
			CreatedAt:        time.Time{},
		}

		sm.On("CreateEvent", context.Background(), event).Return(assert.AnError).Once()
		err := s.ToError(context.Background(), ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("some error"),
		})

		assert.Equal(t, "id: some error", err.Error())

		assertExpectations(t)
	})
}
