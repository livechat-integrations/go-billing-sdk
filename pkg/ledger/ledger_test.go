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
	//TODO implement me
	panic("implement me")
}

func (m *apiMock) CreateRecurrentChargeV2(ctx context.Context, params livechat.CreateRecurrentChargeV2Params) (*livechat.RecurrentChargeV2, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*livechat.RecurrentChargeV2), args.Error(1)
}

func (m *apiMock) CancelRecurrentCharge(ctx context.Context, id string) (*livechat.RecurrentChargeV2, error) {
	//TODO implement me
	panic("implement me")
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

func (m *storageMock) GetChargeByIdAndType(ctx context.Context, ID string, chargeType ChargeType) (*Charge, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetTopUpByIdAndType(ctx context.Context, ID string, topUpType TopUpType) (*TopUp, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetTopUpsByOrganizationID(ctx context.Context, organizationID string) ([]TopUp, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) CreateTopUp(ctx context.Context, it TopUp) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) CreateEvent(ctx context.Context, e Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *storageMock) UpdateTopUpStatus(ctx context.Context, ID string, status TopUpStatus) error {
	//TODO implement me
	panic("implement me")
}

func (m *storageMock) UpdateChargeStatus(ctx context.Context, ID string, status ChargeStatus) error {
	//TODO implement me
	panic("implement me")
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
		domainCharge := Charge{
			ID:               "id",
			LCOrganizationID: lcoid,
			Amount:           amount,
			Type:             ChargeTypeRecurrent,
			Status:           ChargeStatusPending,
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
			Type:           ChargeTypeRecurrent,
			Config: ChargeConfig{
				Months: &months,
			},
		})

		assert.Equal(t, "id", id)
		assert.Nil(t, err)

		assertExpectations(t)
	})

}
