package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

var am = new(apiMock)
var sm = new(storageMock)
var em = new(eventMock)
var xm = new(xIdMock)
var xid = "2341"
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
	mock.AssertExpectationsForObjects(t, am, sm, em, xm)
	am.Calls = nil
	sm.Calls = nil
	em.Calls = nil
	xm.Calls = nil

	am.ExpectedCalls = nil
	sm.ExpectedCalls = nil
	em.ExpectedCalls = nil
	xm.ExpectedCalls = nil
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

func (m *eventMock) CreateEvent(ctx context.Context, e common.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *eventMock) ToError(ctx context.Context, params common.ToErrorParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *eventMock) ToEvent(ctx context.Context, organizationID string, action common.EventAction, eventType common.EventType, payload any) common.Event {
	args := m.Called(ctx, organizationID, action, eventType, payload)
	return args.Get(0).(common.Event)
}

type apiMock struct {
	mock.Mock
}

func (m *apiMock) GetDirectCharge(ctx context.Context, id string) (*livechat.DirectCharge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) GetRecurrentChargeV2(ctx context.Context, id string) (*livechat.RecurrentChargeV2, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) CreateDirectCharge(ctx context.Context, params livechat.CreateDirectChargeParams) (*livechat.DirectCharge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) CreateRecurrentChargeV2(ctx context.Context, params livechat.CreateRecurrentChargeV2Params) (*livechat.RecurrentChargeV2, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) CancelRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentChargeV2, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) GetRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

func (m *apiMock) CreateRecurrentCharge(ctx context.Context, params livechat.CreateRecurrentChargeParams) (*livechat.RecurrentCharge, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentCharge), args.Error(1)
}

type storageMock struct {
	mock.Mock
}

func (m *storageMock) CreateEvent(ctx context.Context, event common.Event) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetChargesByOrganizationID(ctx context.Context, lcID string) ([]Charge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) DeleteCharge(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) DeleteSubscriptionByChargeID(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
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

func (m *storageMock) UpdateChargePayload(ctx context.Context, id string, payload livechat.BaseCharge) error {
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateCharge, common.EventTypeInfo, payload).Return(levent).Once()

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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: "masterOrgID",
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		em.On("ToEvent", context.Background(), "masterOrgID", common.EventActionCreateCharge, common.EventTypeInfo, payload).Return(levent).Once()

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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
			Error:            "error creating recurrent charge",
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
			Error:            "error creating recurrent charge",
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
			Error:            "error creating recurrent charge",
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateSubscription, common.EventTypeInfo, payload).Return(levent).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error plan not found", func(t *testing.T) {
		payload := map[string]interface{}{"planName": "notFound", "chargeID": "xyz"}
		sc, _ := json.Marshal(payload)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateSubscription, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("plan not found"),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "xyz", "notFound")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error getting charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateSubscription, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to get charge by organization id: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error charge is nil", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, nil).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateSubscription, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("charge not found"),
		}).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), lcoid, "id", "super")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error creating subscription", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		xm.On("GenerateId").Return(xid, nil)
		sm.On("CreateSubscription", ctx, mock.Anything).Return(assert.AnError).Once()
		payload := map[string]interface{}{"planName": "super", "chargeID": "id"}
		sc, _ := json.Marshal(payload)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateSubscription,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateSubscription, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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

func TestService_GetPremium(t *testing.T) {
	t.Run("success when no LC charge", func(t *testing.T) {
		rsubs := []Subscription{{
			ID: "id",
		}}
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(rsubs, nil).Once()

		subs, err := s.GetPremium(context.Background(), "id")

		assert.Len(t, subs, 1)
		assert.True(t, subs[0].IsActive())
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("success when LC charge is active", func(t *testing.T) {
		baseCharge := livechat.RecurrentCharge{
			BaseCharge: livechat.BaseCharge{
				ID:     "id",
				Status: "active",
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

		subs, err := s.GetPremium(context.Background(), "id")

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

		subs, err := s.GetPremium(context.Background(), "id")

		assert.Len(t, subs, 0)
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("no subscriptions", func(t *testing.T) {
		var rsubs []Subscription
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(rsubs, nil).Once()

		subs, err := s.GetPremium(context.Background(), "id")

		assert.Len(t, subs, 0)
		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		sm.On("GetSubscriptionsByOrganizationID", ctx, "id").Return(nil, assert.AnError).Once()

		subs, err := s.GetPremium(context.Background(), "id")

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
			payload := args.Get(2).(livechat.BaseCharge)
			assert.NotNil(t, payload)
			assert.Equal(t, "name", payload.Name)
			assert.Equal(t, 10, payload.Price)
		}).Return(nil).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(charge)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionSyncRecurrentCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error getting charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()
		payload := map[string]interface{}{"id": "id"}
		sc, _ := json.Marshal(payload)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionSyncRecurrentCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionSyncRecurrentCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionSyncRecurrentCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncRecurrentCharge,
			Payload:          sc,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionSyncRecurrentCharge, common.EventTypeInfo, payload).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to update charge payload: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}
