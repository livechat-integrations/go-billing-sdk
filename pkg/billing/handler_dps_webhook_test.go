package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
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
	wCtx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
	wCtx = context.WithValue(wCtx, BillingEventIDCtxKey{}, xid)
	wCtx = context.WithValue(wCtx, BillingOrganizationIDCtxKey{}, lcoid)

	return wCtx
}

type billingMock struct {
	mock.Mock
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
			License:          654,
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
		ch2 := Charge{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Type:             ChargeTypeRecurring,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("GetChargesByOrganizationID", billingCtx, lcoid).Return([]Charge{ch1, ch2}, nil).Once()
		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, ch1.ID).Return(nil).Once()
		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, ch2.ID).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookApplicationUninstalled, common.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.Nil(t, err)

		assertExpectations(t)
	})
	t.Run("application_uninstalled get charges error", func(t *testing.T) {
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}

		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("GetChargesByOrganizationID", billingCtx, lcoid).Return(nil, assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookApplicationUninstalled, common.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, common.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
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
			License:          654,
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
		ch2 := Charge{
			ID:               "id2",
			LCOrganizationID: lcoid,
			Type:             ChargeTypeRecurring,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("GetChargesByOrganizationID", billingCtx, lcoid).Return([]Charge{ch1, ch2}, nil).Once()
		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, ch1.ID).Return(nil).Once()
		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, ch2.ID).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookApplicationUninstalled, common.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("delete subscription with charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		bm.On("CreateSubscription", billingCtx, lcoid, paymentID, planName).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		bm.On("CreateSubscription", billingCtx, lcoid, paymentID, planName).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("sync recurrent charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("payment_collected no plan error", func(t *testing.T) {
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		wCtx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, "")
		wCtx = context.WithValue(wCtx, BillingEventIDCtxKey{}, xid)
		wCtx = context.WithValue(wCtx, BillingOrganizationIDCtxKey{}, lcoid)

		bm.On("SyncRecurrentCharge", wCtx, lcoid, paymentID).Return(nil).Once()
		em.On("ToEvent", wCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", wCtx, common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("no plan name found in context"),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, "")
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("payment_collected create subscription error", func(t *testing.T) {
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		bm.On("SyncRecurrentCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		bm.On("CreateSubscription", billingCtx, lcoid, paymentID, planName).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("create subscription: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeInfo,
			Action:           common.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, paymentID).Return(nil).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", billingCtx, levent).Return(nil).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
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
			License:          654,
			LCOrganizationID: lcoid,
			Payload: map[string]interface{}{
				"paymentID": paymentID,
			},
			UserID: userID,
		}
		sc, _ := json.Marshal(req)
		levent := common.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             common.EventTypeError,
			Action:           common.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		bm.On("DeleteSubscriptionWithCharge", billingCtx, lcoid, paymentID).Return(assert.AnError).Once()
		em.On("ToEvent", billingCtx, lcoid, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", billingCtx, common.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("delete subscription with charge: %w", assert.AnError),
		}).Return(assert.AnError).Once()
		lctx := context.WithValue(context.Background(), BillingSubscriptionPlanNameCtxKey{}, planName)
		err := h.HandleDPSWebhook(lctx, req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}
