package ledger

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

var am = new(apiMock)
var sm = new(storageMock)
var xm = new(xIdMock)

var s = Service{
	xIdProvider: xm,
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

func (x *xIdMock) GenerateXId() string {
	args := x.Called()
	return args.Get(0).(string)
}

type apiMock struct {
	mock.Mock
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

func (m *storageMock) GetChargeById(ctx context.Context, ID string) (*Charge, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Charge), args.Error(1)
}

func (m *storageMock) GetTopUpByIdAndType(ctx context.Context, ID string, topUpType TopUpType) (*TopUp, error) {
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
	})
}

func TestService_CreateCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)
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
	})

	t.Run("error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		xid := "2341"
		params := CreateChargeParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
		}

		xm.On("GenerateXId").Return(xid, nil)
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
		}).Return(nil).Once()

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
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

		rawRC, _ := json.Marshal(rc)
		topUp := TopUp{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             TopUpTypeRecurrent,
			Status:           TopUpStatusPending,
			LCCharge:         rawRC,
		}
		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount,
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
	})

	t.Run("error recurrent no months", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		xid := "2341"
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeRecurrent,
			Config:         TopUpConfig{},
		}

		xm.On("GenerateXId").Return(xid, nil)

		sc, _ := json.Marshal(params)

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.Equal(t, "2341: failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge config months is nil", err.Error())

		assertExpectations(t)
	})

	t.Run("error recurrent api error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount,
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
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error recurrent no api charge returned", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

		am.On("CreateRecurrentChargeV2", ctx, livechat.CreateRecurrentChargeV2Params{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount,
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
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.Equal(t, "2341: failed to create top up billing charge: failed to create recurrent charge v2 via lc: charge is nil", err.Error())

		assertExpectations(t)
	})

	t.Run("success direct", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		xid := "2341"
		rc := &livechat.DirectCharge{
			BaseChargeV2: livechat.BaseChargeV2{
				ID:    "id",
				Name:  "name",
				Test:  false,
				Price: amount,
			},
			Quantity: 1,
		}

		xm.On("GenerateXId").Return(xid, nil)

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
			Price:     amount,
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
	})

	t.Run("error direct api error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		xid := "2341"
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeDirect,
			Config:         TopUpConfig{},
		}

		xm.On("GenerateXId").Return(xid, nil)

		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount,
			Test:      false,
		}).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(params)

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error direct no api charge returned", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		xid := "2341"
		params := CreateTopUpRequestParams{
			Test:           false,
			Name:           "name",
			Amount:         amount,
			OrganizationID: lcoid,
			Type:           TopUpTypeDirect,
			Config:         TopUpConfig{},
		}

		xm.On("GenerateXId").Return(xid, nil)

		am.On("CreateDirectCharge", ctx, livechat.CreateDirectChargeParams{
			Name:      "name",
			ReturnURL: "returnURL",
			Price:     amount,
			Test:      false,
		}).Return(nil, nil).Once()

		sc, _ := json.Marshal(params)

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCreateTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		id, err := s.CreateTopUpRequest(context.Background(), params)

		assert.Equal(t, "", id)
		assert.Equal(t, "2341: failed to create top up billing charge: failed to create direct charge via lc: charge is nil", err.Error())

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
	t.Run("success", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

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
		sm.On("GetTopUpByIdAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()
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
	})

	t.Run("top up not found", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)

		sm.On("GetTopUpByIdAndType", ctx, "id", TopUpTypeRecurrent).Return(nil, nil).Once()

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
	})

	t.Run("get top up error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)

		sm.On("GetTopUpByIdAndType", ctx, "id", TopUpTypeRecurrent).Return(nil, assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("api error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

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
		sm.On("GetTopUpByIdAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

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
		sm.On("GetTopUpByIdAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()
		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelTopUpRequest(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("update status not found error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		months := 0
		xid := "2341"
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

		xm.On("GenerateXId").Return(xid, nil)

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
		sm.On("GetTopUpByIdAndType", ctx, "id", TopUpTypeRecurrent).Return(&topUp, nil).Once()
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
	})
}

func TestService_ForceCancelTopUp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)

		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(nil).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": TopUpStatusCancelled})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionForceCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.ForceCancelTopUp(context.Background(), lcoid, "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("not found error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)

		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(ErrNotFound).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "result": "top up not found"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionForceCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.ForceCancelTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, ErrTopUpNotFound)

		assertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)

		sm.On("UpdateTopUpStatus", ctx, "id", TopUpStatusCancelled).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id", "status": TopUpStatusCancelled})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionForceCancelTopUp,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.ForceCancelTopUp(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}

func TestService_CancelCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)
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
	})

	t.Run("charge not found error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)
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
	})

	t.Run("update charge error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		xid := "2341"

		xm.On("GenerateXId").Return(xid, nil)

		sm.On("UpdateChargeStatus", ctx, "id", ChargeStatusCancelled).Return(assert.AnError).Once()

		sc, _ := json.Marshal(map[string]interface{}{"id": "id"})

		sm.On("CreateEvent", ctx, Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionCancelCharge,
			Payload:          sc,
		}).Return(nil).Once()

		err := s.CancelCharge(context.Background(), lcoid, "id")

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}
