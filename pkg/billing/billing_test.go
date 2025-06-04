package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/pkg/events"
)

var am = new(apiMock)
var sm = new(storageMock)
var em = new(eventMock)
var xm = new(xIdMock)
var xid = "2341"
var lid int32 = 654
var lcoid = "lcOrganizationID"

var s = Service{
	billingAPI:   am,
	eventService: em,
	idProvider:   xm,
	storage:      sm,
	plans:        Plans{{Name: "super"}},
	returnURL:    "returnURL",
	masterOrgID:  "masterOrgID",
}
var ctx = context.Background()

var assertExpectations = func(t *testing.T) {
	mock.AssertExpectationsForObjects(t, am, sm, em, xm, bm)
	am.Calls = nil
	sm.Calls = nil
	em.Calls = nil
	xm.Calls = nil
	bm.Calls = nil

	am.ExpectedCalls = nil
	sm.ExpectedCalls = nil
	em.ExpectedCalls = nil
	xm.ExpectedCalls = nil
	bm.ExpectedCalls = nil
}

type xIdMock struct {
	mock.Mock
}

func (x *xIdMock) GenerateId() string {
	args := x.Called()
	return args.Get(0).(string)
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
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

func (m *apiMock) CreateRecurrentChargeV3(ctx context.Context, params livechat.CreateRecurrentChargeParams) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
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

func (m *storageMock) CreateEvent(ctx context.Context, event events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *storageMock) GetChargesByOrganizationID(ctx context.Context, lcID string) ([]Charge, error) {
	args := m.Called(ctx, lcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Charge), args.Error(1)
}

func (m *storageMock) DeleteCharge(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *storageMock) DeleteSubscriptionByChargeID(ctx context.Context, LCOrganizationID string, id string) error {
	args := m.Called(ctx, LCOrganizationID, id)
	return args.Error(0)
}

func (m *storageMock) CreateCharge(ctx context.Context, c Charge) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *storageMock) GetCharge(ctx context.Context, id string) (*Charge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*Charge), args.Error(1)
}

func (m *storageMock) UpdateChargePayload(ctx context.Context, id string, payload json.RawMessage) error {
	args := m.Called(ctx, id, payload)
	return args.Error(0)
}

func (m *storageMock) GetChargeByOrganizationID(ctx context.Context, lcID string) (*Charge, error) {
	args := m.Called(ctx, lcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*Charge), args.Error(1)
}

func (m *storageMock) CreateSubscription(ctx context.Context, subscription Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *storageMock) GetSubscriptionsByOrganizationID(ctx context.Context, lcID string) ([]Subscription, error) {
	args := m.Called(ctx, lcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Subscription), args.Error(1)
}

func (m *storageMock) UpdateSubscriptionNextChargeAt(ctx context.Context, id string, nextChargeAt time.Time) error {
	args := m.Called(ctx, id, nextChargeAt)
	return args.Error(0)
}

func (m *storageMock) GetChargesByStatuses(ctx context.Context, statuses []string) ([]Charge, error) {
	args := m.Called(ctx, statuses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Charge), args.Error(1)
}

func TestNewService(t *testing.T) {
	t.Run("NewService", func(t *testing.T) {
		newService := NewService(nil, nil, nil, "labs", func(ctx context.Context) (string, error) { return "", nil }, &storageMock{}, nil, "returnURL", "masterOrgID")

		assert.NotNil(t, newService)
	})
}

func TestService_CreateRecurrentCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: 10,
			},
			TrialDays: 0,
			Months:    1,
		}

		rawRC, _ := json.Marshal(rc)
		domainCharge := Charge{
			ID:               "id",
			Type:             ChargeTypeRecurring,
			Payload:          rawRC,
			LCOrganizationID: lcoid,
		}
		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     10,
			Test:      false,
			TrialDays: 0,
			Months:    1,
		}).Return(rc, nil).Once()
		sm.On("CreateCharge", ctx, domainCharge).Return(nil).Once()
		payload := map[string]interface{}{"name": "name", "price": 10}
		sc, _ := json.Marshal(domainCharge)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateCharge,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateCharge, events.EventTypeInfo, payload).Return(levent).Once()

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, lcoid)

		assert.Equal(t, "id", id)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success with test", func(t *testing.T) {
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  true,
				Price: 10,
			},
			TrialDays: 0,
			Months:    1,
		}

		rawRC, _ := json.Marshal(rc)
		domainCharge := Charge{
			ID:               "id",
			Type:             ChargeTypeRecurring,
			Payload:          rawRC,
			LCOrganizationID: "masterOrgID",
		}
		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     10,
			Test:      true,
			TrialDays: 0,
			Months:    1,
		}).Return(rc, nil).Once()
		sm.On("CreateCharge", ctx, domainCharge).Return(nil).Once()
		payload := map[string]interface{}{"name": "name", "price": 10}
		sc, _ := json.Marshal(domainCharge)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: "masterOrgID",
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateCharge,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		em.On("ToEvent", context.Background(), "masterOrgID", events.EventActionCreateCharge, events.EventTypeInfo, payload).Return(levent).Once()

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, "masterOrgID")

		assert.Equal(t, "id", id)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error creating recurrent charge", func(t *testing.T) {
		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     10,
			Test:      false,
			TrialDays: 0,
			Months:    1,
		}).Return(nil, assert.AnError).Once()
		payload := map[string]interface{}{"name": "name", "price": 10}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateCharge,
			Payload:          sc,
			Error:            "error creating recurrent charge",
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create recurrent charge via lc: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, lcoid)

		assert.Empty(t, id)
		assert.Error(t, err)

		assertExpectations(t)
	})

	t.Run("error charge is nil", func(t *testing.T) {
		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     10,
			Test:      false,
			TrialDays: 0,
			Months:    1,
		}).Return(nil, nil).Once()
		payload := map[string]interface{}{"name": "name", "price": 10}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateCharge,
			Payload:          sc,
			Error:            "error creating recurrent charge",
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create recurrent charge via lc: charge is nil"),
		}).Return(assert.AnError).Once()

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, lcoid)

		assert.Empty(t, id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error creating charge", func(t *testing.T) {
		rc := &livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: 10,
			},
			TrialDays: 0,
			Months:    1,
		}

		rawRC, _ := json.Marshal(rc)
		domainCharge := Charge{
			ID:               "id",
			Type:             ChargeTypeRecurring,
			Payload:          rawRC,
			LCOrganizationID: lcoid,
		}

		am.On("CreateRecurrentCharge", ctx, livechat.CreateRecurrentChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     10,
			Test:      false,
			TrialDays: 0,
			Months:    1,
		}).Return(rc, nil).Once()
		sm.On("CreateCharge", ctx, domainCharge).Return(assert.AnError).Once()
		payload := map[string]interface{}{"name": "name", "price": 10}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateCharge,
			Payload:          sc,
			Error:            "error creating recurrent charge",
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create charge in database: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, lcoid)

		assert.Empty(t, id)
		assert.Error(t, err)

		assertExpectations(t)
	})

}

func TestService_CreateSubscription(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		charge := Charge{
			ID: "id",
		}
		sm.On("GetCharge", ctx, "id").Return(&charge, nil).Once()
		sm.On("GetSubscriptionsByOrganizationID", ctx, lcoid).Return([]Subscription{}, nil).Once()
		sm.On("CreateSubscription", ctx, mock.Anything).Run(func(args mock.Arguments) {
			argsSub := args.Get(1).(Subscription)
			assert.NotNil(t, argsSub)
			assert.Equal(t, "id", argsSub.Charge.ID)
			assert.Equal(t, "super", argsSub.PlanName)
			assert.Equal(t, lcoid, argsSub.LCOrganizationID)
			assert.NotNil(t, argsSub.ID)
		}).Return(nil).Once()
		xm.On("GenerateId").Return(xid, nil)
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(charge)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateSubscription, events.EventTypeInfo, payload).Return(levent).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success subscription already exists", func(t *testing.T) {
		charge := Charge{
			ID: "id",
		}
		sub := Subscription{
			ID:               "sub1",
			Charge:           &charge,
			LCOrganizationID: lcoid,
			PlanName:         "super",
		}
		sm.On("GetSubscriptionsByOrganizationID", ctx, lcoid).Return([]Subscription{sub}, nil).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateSubscription,
			Payload:          sc,
		}

		afterPayload := map[string]interface{}{"planName": "super", "chargeID": "id", "result": "subscription already exists"}
		asc, _ := json.Marshal(afterPayload)
		afterEvent := levent
		afterEvent.Payload = asc

		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateSubscription, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("CreateEvent", context.Background(), afterEvent).Return(nil).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error plan not found", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, lcoid).Return([]Subscription{}, nil).Once()
		payload := map[string]interface{}{"planName": "notFound", "chargeID": "xyz"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateSubscription, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("plan not found"),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "xyz", "notFound")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error getting charge", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, lcoid).Return([]Subscription{}, nil).Once()
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateSubscription, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to get charge by organization id: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error charge is nil", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, lcoid).Return([]Subscription{}, nil).Once()
		sm.On("GetCharge", ctx, "id").Return(nil, nil).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateSubscription, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("charge not found"),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error creating subscription", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, lcoid).Return([]Subscription{}, nil).Once()
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		xm.On("GenerateId").Return(xid, nil)
		sm.On("CreateSubscription", ctx, mock.Anything).Return(assert.AnError).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionCreateSubscription, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create subscription in database: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_GetCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()

		charge, err := s.GetCharge(context.Background(), "id")

		assert.NotNil(t, charge)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()

		charge, err := s.GetCharge(context.Background(), "id")

		assert.Nil(t, charge)
		assert.Error(t, err)

		assertExpectations(t)
	})
}

func TestService_IsPremium(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return([]Subscription{{
			ID: "id",
		}}, nil).Once()

		premium, err := s.IsPremium(context.Background(), "id")

		assert.True(t, premium)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(nil, assert.AnError).Once()

		premium, err := s.IsPremium(context.Background(), "id")

		assert.False(t, premium)
		assert.Error(t, err)

		assertExpectations(t)
	})

	t.Run("not premium", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(nil, nil).Once()

		premium, err := s.IsPremium(context.Background(), "id")

		assert.False(t, premium)
		assert.Nil(t, err)

		assertExpectations(t)
	})
}

func TestService_GetActiveSubscriptionsByOrganizationID(t *testing.T) {
	t.Run("success when no LC charge", func(t *testing.T) {
		rsubs := []Subscription{{
			ID: "id",
		}}
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(rsubs, nil).Once()

		subs, err := s.GetActiveSubscriptionsByOrganizationID(context.Background(), "id")

		assert.Len(t, subs, 1)
		assert.True(t, subs[0].IsActive())
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("success when LC charge is active", func(t *testing.T) {
		tNextDay := time.Now().AddDate(0, 0, 1)
		baseCharge := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     "id",
				Status: "active",
			},
			NextChargeAt: &tNextDay,
		}
		payload, _ := json.Marshal(baseCharge)
		rsubs := []Subscription{{
			ID: "id",
			Charge: &Charge{
				ID:      "id",
				Type:    ChargeTypeRecurring,
				Payload: payload,
			},
		}}
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(rsubs, nil).Once()

		subs, err := s.GetActiveSubscriptionsByOrganizationID(context.Background(), "id")

		assert.Len(t, subs, 1)
		assert.True(t, subs[0].IsActive())
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("success when LC charge is not active", func(t *testing.T) {
		baseCharge := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     "id",
				Status: "accepted",
			},
		}
		payload, _ := json.Marshal(baseCharge)
		rsubs := []Subscription{{
			ID: "id",
			Charge: &Charge{
				ID:      "id",
				Type:    ChargeTypeRecurring,
				Payload: payload,
			},
		}}
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(rsubs, nil).Once()

		subs, err := s.GetActiveSubscriptionsByOrganizationID(context.Background(), "id")

		assert.Len(t, subs, 0)
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("no subscriptions", func(t *testing.T) {
		var rsubs []Subscription
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(rsubs, nil).Once()

		subs, err := s.GetActiveSubscriptionsByOrganizationID(context.Background(), "id")

		assert.Len(t, subs, 0)
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(nil, assert.AnError).Once()

		subs, err := s.GetActiveSubscriptionsByOrganizationID(context.Background(), "id")

		assert.Nil(t, subs)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_SyncRecurrentCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		charge := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: 10,
			},
			TrialDays: 0,
			Months:    1,
		}
		am.On("GetRecurrentCharge", ctx, "id").Return(&charge, nil).Once()
		sm.On("UpdateChargePayload", ctx, "id", mock.Anything).Run(func(args mock.Arguments) {
			payload := args.Get(2).(json.RawMessage)
			var p livechat.BaseCharge
			_ = json.Unmarshal(payload, &p)
			assert.NotNil(t, payload)
			assert.Equal(t, "name", p.Name)
			assert.Equal(t, 10, p.Price)
		}).Return(nil).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(charge)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error getting charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to get charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error charge is nil", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, nil).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("charge not found"),
		}).Return(assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error getting recurrent charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		am.On("GetRecurrentCharge", ctx, "id").Return(nil, assert.AnError).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to get recurrent charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error updating charge payload", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		am.On("GetRecurrentCharge", ctx, "id").Return(&livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: 10,
			},
			TrialDays: 0,
			Months:    1,
		}, nil).Once()
		sm.On("UpdateChargePayload", ctx, "id", mock.Anything).Return(assert.AnError).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(payload)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to update charge payload: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_SyncCharges(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sm.On("GetChargesByStatuses", ctx, GetSyncValidStatuses()).Return([]Charge{
			{
				ID:               "some-id",
				LCOrganizationID: lcoid,
			},
		}, nil).Once()
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, map[string]interface{}{"id": "some-id"}).Return(events.Event{}).Once()
		am.On("GetRecurrentCharge", ctx, "some-id").Return(&livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{},
		}, nil).Once()

		sm.On("UpdateChargePayload", ctx, "some-id", mock.Anything).Return(nil).Once()
		em.On("CreateEvent", ctx, mock.Anything).Return(nil).Once()

		err := s.SyncCharges(ctx)
		assert.NoError(t, err)

		assertExpectations(t)
	})

	t.Run("error getting charges", func(t *testing.T) {
		sm.On("GetChargesByStatuses", ctx, GetSyncValidStatuses()).Return(nil, errors.New("woopsie")).Once()

		err := s.SyncCharges(ctx)
		assert.ErrorContains(t, err, "failed to get charges by statuses")

		assertExpectations(t)
	})

	t.Run("error getting recurrent charge", func(t *testing.T) {
		sm.On("GetChargesByStatuses", ctx, GetSyncValidStatuses()).Return([]Charge{
			{
				ID:               "some-id",
				LCOrganizationID: lcoid,
			},
		}, nil).Once()
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, map[string]interface{}{"id": "some-id"}).Return(events.Event{}).Once()
		am.On("GetRecurrentCharge", ctx, "some-id").Return(nil, errors.New("whoopsie")).Once()
		em.On("ToError", ctx, mock.Anything).Return(errors.New("failed to get recurrent charge: whoopsie")).Once()
		err := s.SyncCharges(ctx)
		assert.ErrorContains(t, err, "failed to get recurrent charge")

		assertExpectations(t)
	})

	t.Run("error updating payload", func(t *testing.T) {
		sm.On("GetChargesByStatuses", ctx, GetSyncValidStatuses()).Return([]Charge{
			{
				ID:               "some-id",
				LCOrganizationID: lcoid,
			},
		}, nil).Once()
		em.On("ToEvent", ctx, lcoid, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, map[string]interface{}{"id": "some-id"}).Return(events.Event{}).Once()
		am.On("GetRecurrentCharge", ctx, "some-id").Return(&livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{},
		}, nil).Once()

		sm.On("UpdateChargePayload", ctx, "some-id", mock.Anything).Return(errors.New("whoopsie")).Once()
		em.On("ToError", ctx, mock.Anything).Return(errors.New("failed to update charge payload: whoopsie")).Once()

		err := s.SyncCharges(ctx)
		assert.ErrorContains(t, err, "failed to update charge payload")

		assertExpectations(t)
	})
}
