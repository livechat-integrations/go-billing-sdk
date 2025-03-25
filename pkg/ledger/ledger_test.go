package ledger

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

var am = new(apiMock)
var sm = new(storageMock)
var xm = new(xIdMock)

var s = Service{
	idProvider:  xm,
	billingAPI:  am,
	storage:     sm,
	returnURL:   "returnURL",
	masterOrgID: "masterOrgID",
}
var ctx = context.Background()

var assertExpectations = func(t *testing.T) {
	mock.AssertExpectationsForObjects(t, am, sm)
	am.Calls = nil
	sm.Calls = nil

	am.ExpectedCalls = nil
	sm.ExpectedCalls = nil
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

func (m *storageMock) GetTopUpByID(ctx context.Context, ID string) (*TopUp, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
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

func (m *storageMock) GetTopUpByIDAndType(ctx context.Context, ID string, topUpType TopUpType) (*TopUp, error) {
	args := m.Called(ctx, ID, topUpType)
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

func (m *storageMock) CreateTopUp(ctx context.Context, t TopUp) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *storageMock) CreateEvent(ctx context.Context, e Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *storageMock) UpdateTopUpStatus(ctx context.Context, ID string, status TopUpStatus) error {
	args := m.Called(ctx, ID, status)
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
		newService := NewService(nil, nil, "labs", func(ctx context.Context) (string, error) { return "", nil }, &storageMock{}, "returnURL", "masterOrgID")

		assert.NotNil(t, newService)
		assertExpectations(t)
	})
}

func TestService_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"

		call := xm.On("GenerateId").Return(xid, nil)
		domainCharge := Charge{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           amount,
			Status:           ChargeStatusActive,
		}

		sm.On("CreateCharge", ctx, domainCharge).Return(nil).Once()
		sc, _ := json.Marshal(domainCharge)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCreateCharge,
			Payload:          sc,
		}).Return(nil).Once()

		id, err := s.CreateCharge(context.Background(), CreateChargeParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
		})

		assert.Equal(t, xid, id)
		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		domainCharge := Charge{
			ID:               xid,
			LCOrganizationID: lcoid,
			Amount:           amount,
			Status:           ChargeStatusActive,
		}

		sm.On("CreateCharge", ctx, domainCharge).Return(assert.AnError).Once()
		sc, _ := json.Marshal(params)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateCharge,
			Payload:          sc,
			Error:            "failed to create charge in database: assert.AnError general error for testing",
		}).Return(nil).Once()

		id, err := s.CreateCharge(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)

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
		sm.On("CreateTopUp", ctx, topUp).Return(nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeRecurrent,
			Config: TopUpConfig{
				Months: &months,
			},
		})

		assert.Equal(t, "id", id)
		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		sc, _ := json.Marshal(params)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge config months is nil",
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.Equal(t, "2341: failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge config months is nil", err.Error())

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge v2 via lc: assert.AnError general error for testing",
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
			TrialDays: 0,
			Months:    months,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge is nil",
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.Equal(t, "2341: failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge is nil", err.Error())

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
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
		sm.On("CreateTopUp", ctx, topUp).Return(nil).Once()

		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeDirect,
			Config:         TopUpConfig{},
		})

		assert.Equal(t, "id", id)
		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create direct charge via lc: assert.AnError general error for testing",
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount * 100,
			Test:      false,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
			Error:            "failed to create top up billing charge: failed to create direct charge via lc: charge is nil",
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.Equal(t, "2341: failed to create top up billing charge: failed to create direct charge via lc: charge is nil", err.Error())

		assertExpectations(t)
		call.Unset()
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
	t.Run("success", func(t *testing.T) {
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

		call := xm.On("GenerateId").Return(xid, nil)

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
		sm.On("GetTopUpByIDAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "success"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("top up not found", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("GetTopUpByIDAndType", ctx, "id", TopUpTypeRecurrent).Return(nil, nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("get top up error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("GetTopUpByIDAndType", ctx, "id", TopUpTypeRecurrent).Return(nil, assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
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
		call := xm.On("GenerateId").Return(xid, nil)

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
		sm.On("GetTopUpByIDAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("update status error", func(t *testing.T) {
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
		call := xm.On("GenerateId").Return(xid, nil)

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
		sm.On("GetTopUpByIDAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("update status not found error", func(t *testing.T) {
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
		call := xm.On("GenerateId").Return(xid, nil)

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
		sm.On("GetTopUpByIDAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(ErrNotFound).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
		call.Unset()
	})
}

func TestService_UpdateTopUpStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("UpdateTopUpStatus", ctx, "id", status).Return(nil).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": status})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionUpdateTopUpStatus,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.UpdateTopUpStatus(context.Background(), lcoid, "id", status)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("not found error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("UpdateTopUpStatus", ctx, "id", status).Return(ErrNotFound).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionUpdateTopUpStatus,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.UpdateTopUpStatus(context.Background(), lcoid, "id", status)

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		status := TopUpStatusCancelled

		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("UpdateTopUpStatus", ctx, "id", status).Return(assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": status})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionUpdateTopUpStatus,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}).Return(nil).Once()

		err := s.UpdateTopUpStatus(context.Background(), lcoid, "id", status)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})
}

func TestService_CancelCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "success"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCancelCharge,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("charge not found error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(ErrNotFound).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "charge not found"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionCancelCharge,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrChargeNotFound)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("update charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		call := xm.On("GenerateId").Return(xid, nil)
		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelCharge,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}).Return(nil).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})
}

func TestService_SyncTopUp(t *testing.T) {
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
				Price:             amount,
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
			Status:           TopUpStatusCancelled,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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
			Status:           TopUpStatusCancelled,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
			Error:            "charge conflict",
		}).Return(nil).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.Equal(t, "2341: charge conflict", err.Error())

		assertExpectations(t)
		call.Unset()
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(nil, assert.AnError).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(&rc, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
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
				Price:             amount,
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

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, assert.AnError).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(&topUp, nil).Once()
		sc, _ := json.Marshal(topUp)
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
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
		call.Unset()
	})

	t.Run("no api charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(nil, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
			Error:            "charge not found",
		}).Return(nil).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.Equal(t, "2341: charge not found", err.Error())

		assertExpectations(t)
		call.Unset()
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
				Price:             amount,
				ReturnURL:         "utp://lorem$%^#09sd90 url",
				Test:              false,
				PerAccount:        false,
				Status:            "failed",
				ConfirmationURL:   confUrl,
				CommissionPercent: 10,
			},
			Quantity: 1,
		}

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
			Error:            "parse \"utp://lorem$%^\": invalid URL escape \"%^\"",
		}).Return(nil).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.Error(t, err)

		assertExpectations(t)
		call.Unset()
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
				Price:             amount,
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
			Status:           TopUpStatusCancelled,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  confUrl,
			LCCharge:         jdc,
		}

		call := xm.On("GenerateId").Return(xid, nil)
		am.On("GetDirectCharge", ctx, "id").Return(&dc, nil).Once()
		am.On("GetRecurrentChargeV2", ctx, "id").Return(nil, nil).Once()
		sm.On("UpsertTopUp", ctx, topUp).Return(nil, assert.AnError).Once()
		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})
		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionSyncTopUp,
			Payload:          sc,
			Error:            "assert.AnError general error for testing",
		}).Return(nil).Once()

		_, err := s.SyncTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})
}

func TestService_ToEvent(t *testing.T) {
	t.Run("success id from context", func(t *testing.T) {
		id := "id-from-context"
		localCtx := context.WithValue(context.Background(), LedgerEventIDCtxKey{}, id)
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

		call := xm.On("GenerateId").Return(id, nil)
		event := s.ToEvent(context.Background(), lcoid, action, eventType, payload)

		assert.Equal(t, eventType, event.Type)
		assert.Equal(t, id, event.ID)
		assert.Equal(t, action, event.Action)
		assert.Equal(t, lcoid, event.LCOrganizationID)

		assertExpectations(t)
		call.Unset()
	})
}
