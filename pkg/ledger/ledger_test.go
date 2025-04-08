package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
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

func (m *apiMock) GetRecurrentChargeV2(ctx context.Context, id string) (*livechat.RecurrentChargeV2, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*livechat.RecurrentChargeV2), args.Error(1)
}

func (m *apiMock) CreateDirectCharge(ctx context.Context, params livechat.CreateDirectChargeParams) (*livechat.DirectCharge, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.DirectCharge), args.Error(1)
}

func (m *apiMock) CreateRecurrentChargeV2(ctx context.Context, params livechat.CreateRecurrentChargeV2Params) (*livechat.RecurrentChargeV2, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentChargeV2), args.Error(1)
}

func (m *apiMock) CancelRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentChargeV2, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*livechat.RecurrentChargeV2), args.Error(1)
}

func (m *apiMock) GetRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentCharge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) CreateRecurrentCharge(ctx context.Context, params livechat.CreateRecurrentChargeParams) (*livechat.RecurrentCharge, error) {
	//TODO implement me
	panic("implement me")
}

type storageMock struct {
	mock.Mock
}

func (m *storageMock) CreateEvent(ctx context.Context, event common.Event) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) InitRecurrentTopUpRequiredValues(ctx context.Context, params InitRecurrentTopUpRequiredValuesParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
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

func (m *storageMock) GetChargeById(ctx context.Context, ID string) (*Charge, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Charge), args.Error(1)
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

func (m *storageMock) UpdateChargeStatus(ctx context.Context, ID string, status ChargeStatus) error {
	args := m.Called(ctx, ID, status)
	return args.Error(0)
}

func (m *storageMock) CreateCharge(ctx context.Context, c Charge) error {
	args := m.Called(ctx, c)
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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"

		xm.On("GenerateId").Return(xid, nil)
		domainCharge := Charge{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           amount,
			Status:           ChargeStatusActive,
		}

		sm.On("CreateCharge", ctx, domainCharge).Return(nil).Once()
		sc, _ := json.Marshal(domainCharge)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
		}
		em.On("CreateEvent", context.Background(), levent).Return(nil).Once()
		params := CreateChargeParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateCharge, common.EventTypeInfo, params).Return(levent).Once()

		id, err := s.CreateCharge(context.Background(), params)

		assert.Equal(t, xid, id)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		params := CreateChargeParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
		}

		xm.On("GenerateId").Return(xid, nil)
		domainCharge := Charge{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           amount,
			Status:           ChargeStatusActive,
		}

		sm.On("CreateCharge", ctx, domainCharge).Return(assert.AnError).Once()
		sc, _ := json.Marshal(params)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateCharge,
			Payload:          sc,
			Error:            "failed to create charge in database: assert.AnError general error for testing",
		}
		em.On("ToEvent", context.Background(), lcoid, common.EventActionCreateCharge, common.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create charge in database: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		id, err := s.CreateCharge(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_CreateTopUpRequest(t *testing.T) {
	t.Run("success recurrent", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		confUrl := "http://livechat.com/confirmation"
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:              "id",
				Name:            "name",
				Test:            false,
				Price:           amount,
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
		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCreateTopUp,
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
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()

		tu, err := s.CreateTopUpRequest(ctx, params)

		assert.Equal(t, &topUp, tu)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error recurrent no months", func(t *testing.T) {
		amount := float32(5.234)
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge config months is nil",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create recurrent charge v2 via lc: charge config months is nil")),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error recurrent api error", func(t *testing.T) {
		amount := float32(5.234)
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

		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge v2 via lc: assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create recurrent charge v2 via lc: %w", assert.AnError)),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error recurrent no api charge returned", func(t *testing.T) {
		amount := float32(5.234)
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

		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge v2 via lc: ",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create recurrent charge v2 via lc: charge is nil")),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success direct", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		rc := &livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
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
			Price:     amount * 100,
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCreateTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, &topUp, tu)
		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error direct api error", func(t *testing.T) {
		amount := float32(5.234)
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
			Price:     amount * 100,
			Test:      false,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create direct charge via lc: assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", fmt.Errorf("failed to create direct charge via lc: %w", assert.AnError)),
		}).Return(assert.AnError).Once()

		tu, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Nil(t, tu)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error direct no api charge returned", func(t *testing.T) {
		amount := float32(5.234)
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
			Price:     amount * 100,
			Test:      false,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create direct charge via lc: charge is nil",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCreateTopUp, common.EventTypeInfo, params).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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
		amount := float32(5.234)
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
	t.Run("success without current topped up at", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
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

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success with current topped up at", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusPending,
			LCCharge:          rawRC,
			CurrentToppedUpAt: &someDate,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(rc, nil).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:                "id",
			Status:            TopUpStatusCancelled,
			CurrentToppedUpAt: &someDate,
		}).Return(nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "success"})

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
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

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("api error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
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

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status error without current topped up at", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
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

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status error with current topped up at", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		someDate, _ := time.Parse("2006-01-02", "1999-01-01")
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusPending,
			LCCharge:          rawRC,
			CurrentToppedUpAt: &someDate,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(rc, nil).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:                "id",
			Status:            TopUpStatusCancelled,
			CurrentToppedUpAt: &someDate,
		}).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status not found error without current topped up at", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
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

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
	})

	t.Run("update status not found error with current topped up at", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		someDate, _ := time.Parse("2006-01-02", "2021-01-01")
		rc := &livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
			},
			TrialDays: 0,
			Months:    months,
		}

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			Status:            TopUpStatusPending,
			LCCharge:          rawRC,
			CurrentToppedUpAt: &someDate,
		}
		am.On("CancelRecurrentCharge", ctx, "id").Return(rc, nil).Once()
		sm.On("GetTopUpByIDAndType", ctx, GetTopUpByIDAndTypeParams{
			ID:   "id",
			Type: TopUpTypeRecurrent,
		}).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:                "id",
			Status:            TopUpStatusCancelled,
			CurrentToppedUpAt: &someDate,
		}).Return(ErrNotFound).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
	})
}

func TestService_ForceCancelTopUp(t *testing.T) {
	t.Run("success without current topped up at", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: status,
		}).Return(nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": status})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
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

	t.Run("success with current topped up at", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:                "id",
			Status:            status,
			CurrentToppedUpAt: &someDate,
		}).Return(nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": status})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            status,
			CurrentToppedUpAt: &someDate,
		}

		err := s.ForceCancelTopUp(context.Background(), topUp)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("not found error without current topped up at", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id",
			Status: status,
		}).Return(ErrNotFound).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
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

	t.Run("not found error with current topped up at", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")

		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:                "id",
			Status:            status,
			CurrentToppedUpAt: &someDate,
		}).Return(ErrNotFound).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		topUp := TopUp{
			ID:                "id",
			LCOrganizationID:  lcoid,
			Status:            status,
			CurrentToppedUpAt: &someDate,
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
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id", "status": TopUpStatusCancelled}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
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

func TestService_CancelCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "success"})

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("charge not found error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(ErrNotFound).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "charge not found"})

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionCancelCharge,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrChargeNotFound)

		assertExpectations(t)
	})

	t.Run("update charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionCancelCharge,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_SyncTopUp(t *testing.T) {
	t.Run("success direct cancelled", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			Status:           TopUpStatusProcessing,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			Status:            TopUpStatusProcessing,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc,
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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

	t.Run("error recurrent and direct", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "failed",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount,
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

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
			Error:            "charge conflict",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("charge conflict"),
		}).Return(assert.AnError).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("direct api error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, assert.AnError).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "pending",
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, assert.AnError).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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

	t.Run("no api charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
			Error:            "charge not found",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("charge not found"),
		}).Return(assert.AnError).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("url parse error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
				ReturnURL:         "utp://lorem$%^#09sd90 url",
				Test:              false,
				PerAccount:        false,
				Status:            "failed",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
			Error:            "parse \"utp://lorem$%^\": invalid URL escape \"%^\"",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err: &url.Error{
				Op:  "parse",
				URL: "utp://lorem$%^",
				Err: url.EscapeError("%^"),
			},
		}).Return(assert.AnError).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("upsert error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"

		dc := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(nil, assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error init recurrent top up values", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(assert.AnError).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.Nil(t, tp)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success on init recurrent top up values row not found", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id"}).Return(levent).Once()
		em.On("CreateEvent", ctx, levent).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id",
		}).Return(ErrNotFound).Once()

		tp, err := s.SyncTopUp(context.Background(), lcoid, "id")

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
}

func TestService_SyncOrCancelAllPendingTopUpRequests(t *testing.T) {
	t.Run("success recurrent and direct active", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc1 := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetTopUpsByOrganizationIDAndStatus", ctx, lcoid, TopUpStatusPending).Return([]TopUp{topUp1, topUp2}, nil).Once()
		am.On("GetDirectCharge", ctx, "id1").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id1").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", ctx, "id2").Return(&rc2, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id2").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp2).Return(&topUp2, nil).Once()
		sc1, _ := json.Marshal(topUp1)
		sc2, _ := json.Marshal(topUp2)
		levent1 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id1"}).Return(levent1).Once()
		em.On("CreateEvent", ctx, levent1).Return(nil).Once()
		levent2 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id2"}).Return(levent2).Once()
		em.On("CreateEvent", ctx, levent2).Return(nil).Once()

		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id1",
		}).Return(nil).Once()

		err := s.SyncOrCancelAllPendingTopUpRequests(context.Background(), lcoid)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success recurrent and direct pending", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		now := time.Now()

		rc1 := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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

		oTopUp1 := TopUp{
			ID:                "id1",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusPending,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   confUrl,
			LCCharge:          jrc1,
			CurrentToppedUpAt: &now,
			NextTopUpAt:       &someDate2,
		}
		topUp1 := oTopUp1
		topUp1.CreatedAt = now
		topUp1.UpdatedAt = now

		oTopUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}
		topUp2 := oTopUp2
		topUp2.CreatedAt = now
		topUp2.UpdatedAt = now

		sm.On("GetTopUpsByOrganizationIDAndStatus", ctx, lcoid, TopUpStatusPending).Return([]TopUp{topUp1, topUp2}, nil).Once()
		am.On("GetDirectCharge", ctx, "id1").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id1").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", ctx, "id2").Return(&rc2, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id2").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, oTopUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", ctx, oTopUp2).Return(&topUp2, nil).Once()
		sc1, _ := json.Marshal(oTopUp1)
		sc2, _ := json.Marshal(oTopUp2)
		levent1 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id1"}).Return(levent1).Once()
		em.On("CreateEvent", ctx, levent1).Return(nil).Once()
		levent2 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id2"}).Return(levent2).Once()
		em.On("CreateEvent", ctx, levent2).Return(nil).Once()
		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: now,
			NextTopUpAt:       someDate2,
			ID:                "id1",
		}).Return(nil).Once()

		err := s.SyncOrCancelAllPendingTopUpRequests(context.Background(), lcoid)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("force cancel all old pending", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc1 := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
		rc2 := livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
				ReturnURL:         "http://www.google.com",
				Test:              false,
				PerAccount:        false,
				Status:            "other",
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

		topUp2 := TopUp{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusPending,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetTopUpsByOrganizationIDAndStatus", ctx, lcoid, TopUpStatusPending).Return([]TopUp{topUp1, topUp2}, nil).Once()
		am.On("GetDirectCharge", ctx, "id1").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id1").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", ctx, "id2").Return(&rc2, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id2").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp1).Return(&topUp1, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp2).Return(&topUp2, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:                "id1",
			Status:            TopUpStatusCancelled,
			CurrentToppedUpAt: &someDate,
		}).Return(nil).Once()
		sm.On("UpdateTopUpStatus", ctx, UpdateTopUpStatusParams{
			ID:     "id2",
			Status: TopUpStatusCancelled,
		}).Return(nil).Once()

		sc1, _ := json.Marshal(topUp1)
		sc2, _ := json.Marshal(topUp2)
		sc11, _ := json.Marshal(map[string]interface{}{"id": "id1", "status": TopUpStatusCancelled})
		sc22, _ := json.Marshal(map[string]interface{}{"id": "id2", "status": TopUpStatusCancelled})

		levent1 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id1"}).Return(levent1).Once()
		em.On("CreateEvent", ctx, levent1).Return(nil).Once()
		levent2 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc2,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id2"}).Return(levent2).Once()
		em.On("CreateEvent", ctx, levent2).Return(nil).Once()

		levent11 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc11,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id1", "status": TopUpStatusCancelled}).Return(levent11).Once()
		em.On("CreateEvent", ctx, levent11).Return(nil).Once()
		levent22 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionForceCancelCharge,
			Payload:          sc22,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionForceCancelCharge, common.EventTypeInfo, map[string]interface{}{"id": "id2", "status": TopUpStatusCancelled}).Return(levent22).Once()
		em.On("CreateEvent", ctx, levent22).Return(nil).Once()

		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id1",
		}).Return(nil).Once()

		err := s.SyncOrCancelAllPendingTopUpRequests(context.Background(), lcoid)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success recurrent and error direct", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		confUrl := "http://www.google.com/confirmation"
		months := 1
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")

		rc1 := livechat.RecurrentChargeV2{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id1",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			BaseChargeV2: livechat.BaseChargeV2{
				ID:                "id2",
				BuyerLicenseID:    123,
				BuyerEntityID:     "321",
				SellerClientID:    "213",
				OrderClientID:     "123",
				OrderLicenseID:    "123",
				OrderEntityID:     "123",
				Name:              "some",
				Price:             amount * 100,
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
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jrc2,
		}

		sm.On("GetTopUpsByOrganizationIDAndStatus", ctx, lcoid, TopUpStatusPending).Return([]TopUp{topUp1, topUp2}, nil).Once()
		am.On("GetDirectCharge", ctx, "id1").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id1").Return(&rc1, nil).Once()
		am.On("GetDirectCharge", ctx, "id2").Return(nil, assert.AnError).Once()
		am.On("GetRecurrentChargeV2", ctx, "id2").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp1).Return(&topUp1, nil).Once()
		sc1, _ := json.Marshal(topUp1)
		sc2, _ := json.Marshal(map[string]interface{}{"id": "id2"})

		levent1 := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc1,
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id1"}).Return(levent1).Once()
		em.On("CreateEvent", ctx, levent1).Return(nil).Once()

		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionSyncTopUp,
			Payload:          sc2,
			Error:            "assert.AnError general error for testing",
		}
		em.On("ToEvent", ctx, lcoid, common.EventActionSyncTopUp, common.EventTypeInfo, map[string]interface{}{"id": "id2"}).Return(levent).Once()
		em.On("ToError", context.Background(), common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		sm.On("InitRecurrentTopUpRequiredValues", ctx, InitRecurrentTopUpRequiredValuesParams{
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
			ID:                "id1",
		}).Return(nil).Once()

		err := s.SyncOrCancelAllPendingTopUpRequests(context.Background(), lcoid)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}
