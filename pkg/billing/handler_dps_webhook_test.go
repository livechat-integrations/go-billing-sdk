package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

var bm = new(billingMock)

var h = Handler{
	eventService: em,
	billing:      bm,
	idProvider:   xm,
}

var planName = "some_plan"
var billingCtx = initCtx()

func initCtx() context.Context {
	wCtx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
	wCtx = context.WithValue(wCtx, EventIDCtxKey{}, xid)
	wCtx = context.WithValue(wCtx, OrganizationIDCtxKey{}, lcoid)
	wCtx = context.WithValue(wCtx, LicenseIDCtxKey{}, lid)

	return wCtx
}

type billingMock struct {
	mock.Mock
}

func (b *billingMock) DeleteSubscription(ctx context.Context, lcOrganizationID string, subscriptionID string) error {
	args := b.Called(ctx, lcOrganizationID, subscriptionID)
	return args.Error(0)
}

func (b *billingMock) SyncCharges(ctx context.Context) error {
	args := b.Called(ctx)
	return args.Error(0)
}

func (b *billingMock) DeleteSubscriptionWithCharge(ctx context.Context, lcOrganizationID string, chargeID string) error {
	args := b.Called(ctx, lcOrganizationID, chargeID)
	return args.Error(0)
}

func (b *billingMock) SyncRecurrentCharge(ctx context.Context, lcOrganizationID string, id string) error {
	args := b.Called(ctx, lcOrganizationID, id)
	return args.Error(0)
}

func (b *billingMock) CreateSubscription(ctx context.Context, lcOrganizationID string, chargeID string, planName string) error {
	args := b.Called(ctx, lcOrganizationID, chargeID, planName)
	return args.Error(0)
}

func (b *billingMock) CreateRecurrentCharge(ctx context.Context, name string, price int, lcOrganizationID string, chargeFrequency int) (string, error) {
	args := b.Called(ctx, name, price, lcOrganizationID, chargeFrequency)
	return args.String(0), args.Error(1)
}

func (b *billingMock) GetChargesByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Charge, error) {
	args := b.Called(ctx, lcOrganizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Charge), nil
}

func (b *billingMock) GetActiveSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error) {
	args := b.Called(ctx, lcOrganizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Subscription), nil
}

func (b *billingMock) GetSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error) {
	args := b.Called(ctx, lcOrganizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]Subscription), nil
}

func (b *billingMock) CreateRecurrentChargeWithTrial(ctx context.Context, name string, price int, lcOrganizationID string, chargeFrequency int) (string, error) {
	args := b.Called(ctx, name, price, lcOrganizationID, chargeFrequency)
	return args.String(0), args.Error(1)
}

func (b *billingMock) HasUsedTrial(ctx context.Context, lcOrganizationID string) (bool, error) {
	args := b.Called(ctx, lcOrganizationID)
	return args.Bool(0), args.Error(1)
}

func TestNewHandler(t *testing.T) {
	t.Run("NewHandler", func(t *testing.T) {
		newService := NewHandler(&eventMock{}, &billingMock{}, &xIdMock{})

		assert.NotNil(t, newService)
		assertExpectations(t)
	})
}

func TestService_HandleDPSWebhook(t *testing.T) {
	t.Run("success application_uninstalled", func(t *testing.T) {
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		ch1 := Charge{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Type:             ChargeTypeRecurring,
		}
		sub := Subscription{
			ID:               "sub1",
			Charge:           &ch1,
			LCOrganizationID: lcoid,
			PlanName:         "unlimited",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("GetSubscriptionsByOrganizationID", billingCtx, lcoid).Return([]Subscription{sub}, nil)
		bm.On("DeleteSubscription", billingCtx, lcoid, sub.ID).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("application_uninstalled get subscriptions error", func(t *testing.T) {
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}

		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("GetSubscriptionsByOrganizationID", billingCtx, lcoid).Return(nil, assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
	t.Run("application_uninstalled delete error", func(t *testing.T) {
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		ch1 := Charge{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Type:             ChargeTypeRecurring,
		}
		sub := Subscription{
			ID:               "sub1",
			Charge:           &ch1,
			LCOrganizationID: lcoid,
			PlanName:         "unlimited",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("GetSubscriptionsByOrganizationID", billingCtx, lcoid).Return([]Subscription{sub}, nil)
		bm.On("DeleteSubscription", billingCtx, lcoid, sub.ID).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("delete subscription with charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success payment_activated", func(t *testing.T) {
		eventType := "payment_activated"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		bm.On("CreateSubscription", billingCtx, lcoid, paymentID, planName).Return(nil).Once()
		bm.On("GetSubscriptionsByOrganizationID", billingCtx, lcoid).Return([]Subscription{}, nil)
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success payment_collected", func(t *testing.T) {
		eventType := "payment_collected"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("GetSubscriptionsByOrganizationID", billingCtx, lcoid).Return([]Subscription{
			{
				ID: "sub1",
			},
		}, nil)
		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("payment_collected sync error", func(t *testing.T) {
		eventType := "payment_collected"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("sync recurrent charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("payment_activated no plan error", func(t *testing.T) {
		eventType := "payment_activated"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		wCtx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, "")
		wCtx = context.WithValue(wCtx, EventIDCtxKey{}, xid)
		wCtx = context.WithValue(wCtx, OrganizationIDCtxKey{}, lcoid)
		wCtx = context.WithValue(wCtx, LicenseIDCtxKey{}, lid)

		bm.On("SyncRecurrentCharge", wCtx, lcoid, paymentID).Return(nil).Once()
		bm.On("GetSubscriptionsByOrganizationID", wCtx, lcoid).Return([]Subscription{}, nil)
		em.On("ToEvent", wCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", wCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("no plan name found in context"),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, "")
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success payment_cancelled", func(t *testing.T) {
		eventType := "payment_cancelled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("payment_cancelled error", func(t *testing.T) {
		eventType := "payment_cancelled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		xm.On("GenerateId").Return(xid, nil)

		req := DPSWebhookRequest{
			ApplicationID:    "123",
			ApplicationName:  "ttt",
			ClientID:         "321",
			Date:             someDate,
			Event:            eventType,
			License:          lid,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, paymentID).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, events.EventActionUnknown, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("delete subscription with charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), SubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}
