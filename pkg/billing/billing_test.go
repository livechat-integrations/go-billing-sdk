package billing

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

var s = Service{
	billingAPI:  am,
	storage:     sm,
	plans:       Plans{{Name: "super"}},
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
		newService := NewService(nil, "labs", func(ctx context.Context) (string, error) { return "", nil }, &storageMock{}, nil, "returnURL", "masterOrgID")

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
			LCOrganizationID: "lcOrganizationID",
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

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, "lcOrganizationID")

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

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, "lcOrganizationID")

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

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, "lcOrganizationID")

		assert.Empty(t, id)
		assert.ErrorContains(t, err, "charge is nil")

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
			LCOrganizationID: "lcOrganizationID",
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

		id, err := s.CreateRecurrentCharge(context.Background(), "name", 10, "lcOrganizationID")

		assert.Empty(t, id)
		assert.Error(t, err)

		assertExpectations(t)
	})

}

func TestService_CreateSubscription(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		sm.On("CreateSubscription", ctx, mock.Anything).Run(func(args mock.Arguments) {
			argsSub := args.Get(1).(Subscription)
			assert.NotNil(t, argsSub)
			assert.Equal(t, "id", argsSub.Charge.ID)
			assert.Equal(t, "super", argsSub.PlanName)
			assert.Equal(t, "lcOrganizationID", argsSub.LCOrganizationID)
			assert.NotNil(t, argsSub.ID)
		}).Return(nil).Once()

		err := s.CreateSubscription(context.Background(), "lcOrganizationID", "id", "super")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error plan not found", func(t *testing.T) {
		err := s.CreateSubscription(context.Background(), "lcOrganizationID", "xyz", "notFound")

		assert.ErrorContains(t, err, "plan not found")

		assertExpectations(t)
	})

	t.Run("error getting charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), "lcOrganizationID", "id", "super")

		assert.Error(t, err)

		assertExpectations(t)
	})

	t.Run("error charge is nil", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, nil).Once()

		err := s.CreateSubscription(context.Background(), "lcOrganizationID", "id", "super")

		assert.ErrorContains(t, err, "charge not found")

		assertExpectations(t)
	})

	t.Run("error creating subscription", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		sm.On("CreateSubscription", ctx, mock.Anything).Return(assert.AnError).Once()

		err := s.CreateSubscription(context.Background(), "lcOrganizationID", "id", "super")

		assert.Error(t, err)

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

func TestService_SyncRecurrentCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
		sm.On("UpdateChargePayload", ctx, "id", mock.Anything).Run(func(args mock.Arguments) {
			payload := args.Get(2).(livechat.BaseCharge)
			assert.NotNil(t, payload)
			assert.Equal(t, "name", payload.Name)
			assert.Equal(t, 10, payload.Price)
		}).Return(nil).Once()

		err := s.SyncRecurrentCharge(context.Background(), "id")

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("error getting charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), "id")

		assert.Error(t, err)

		assertExpectations(t)
	})

	t.Run("error charge is nil", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(nil, nil).Once()

		err := s.SyncRecurrentCharge(context.Background(), "id")

		assert.ErrorContains(t, err, "charge not found")

		assertExpectations(t)
	})

	t.Run("error getting recurrent charge", func(t *testing.T) {
		sm.On("GetCharge", ctx, "id").Return(&Charge{
			ID: "id",
		}, nil).Once()
		am.On("GetRecurrentCharge", ctx, "id").Return(nil, assert.AnError).Once()

		err := s.SyncRecurrentCharge(context.Background(), "id")

		assert.Error(t, err)

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

		err := s.SyncRecurrentCharge(context.Background(), "id")

		assert.Error(t, err)

		assertExpectations(t)
	})
}
