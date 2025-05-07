package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/livechat-integrations/go-billing-sdk/pkg/events"
)

var lm = new(ledgerMock)

var h = Handler{
	eventService: em,
	ledger:       lm,
	idProvider:   xm,
}

var lcoid = "lcOrganizationID"
var xid = "2341"
var ledgerCtx = initCtx()

func initCtx() context.Context {
	ledgerCtx := context.WithValue(context.Background(), LedgerEventIDCtxKey{}, xid)
	ledgerCtx = context.WithValue(ledgerCtx, LedgerOrganizationIDCtxKey{}, lcoid)

	return ledgerCtx
}

type ledgerMock struct {
	mock.Mock
}

func (l *ledgerMock) GetTopUpByIDAndOrganizationID(ctx context.Context, organizationID string, ID string) (*TopUp, error) {
	args := l.Called(ctx, organizationID, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func (l *ledgerMock) SyncTopUp(ctx context.Context, topUp TopUp) (*TopUp, error) {
	args := l.Called(ctx, topUp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func (l *ledgerMock) GetOperations(ctx context.Context, organizationID string) ([]Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) TopUp(ctx context.Context, topUp TopUp) (string, error) {
	args := l.Called(ctx, topUp)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (l *ledgerMock) SyncOrCancelTopUpRequests(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (*TopUp, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) ForceCancelTopUp(ctx context.Context, topUp TopUp) error {
	args := l.Called(ctx, topUp)
	return args.Error(0)
}

func (l *ledgerMock) CreateCharge(ctx context.Context, params CreateChargeParams) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) GetTopUps(ctx context.Context, organizationID string) ([]TopUp, error) {
	args := l.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TopUp), args.Error(1)
}

func (l *ledgerMock) GetTopUpByID(ctx context.Context, ID string) (*TopUp, error) {
	args := l.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func (l *ledgerMock) CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error) {
	args := l.Called(ctx, organizationID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TopUp), args.Error(1)
}

func (l *ledgerMock) UpdateTopUpStatus(ctx context.Context, organizationID string, ID string, status TopUpStatus) error {
	//TODO implement me
	panic("implement me")
}

func TestNewHandler(t *testing.T) {
	t.Run("NewHandler", func(t *testing.T) {
		newService := NewHandler(&eventMock{}, &ledgerMock{}, &xIdMock{})

		assert.NotNil(t, newService)
		assertExpectations(t)
	})
}

func TestService_HandleDPSWebhook(t *testing.T) {
	t.Run("success application_uninstalled", func(t *testing.T) {
		amount := float32(5.234)
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		top2 := TopUp{
			ID:                "id2",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusActive,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   "url",
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return([]TopUp{top1, top2}, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, top1).Return(nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, top2).Return(nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookApplicationUninstalled, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success application_uninstalled no top ups", func(t *testing.T) {
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
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return([]TopUp{}, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookApplicationUninstalled, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("application_uninstalled get top ups error", func(t *testing.T) {
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
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return(nil, assert.AnError).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookApplicationUninstalled, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("application_uninstalled force cancel error", func(t *testing.T) {
		amount := float32(5.234)
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		top2 := TopUp{
			ID:                "id2",
			LCOrganizationID:  lcoid,
			Status:            TopUpStatusActive,
			Amount:            amount,
			Type:              TopUpTypeRecurrent,
			ConfirmationUrl:   "url",
			CurrentToppedUpAt: &someDate,
			NextTopUpAt:       &someDate2,
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return([]TopUp{top1, top2}, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, top1).Return(nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, top2).Return(assert.AnError).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookApplicationUninstalled, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success payment_collected", func(t *testing.T) {
		amount := float32(5.234)
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, top1).Return(&top1, nil).Once()
		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("TopUp", ledgerCtx, top1).Return(top1.ID, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("payment_collected wrong payment id", func(t *testing.T) {
		eventType := "payment_collected"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := float32(234.3)
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
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("payment id field not found in payload"),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("payment_collected get top up error", func(t *testing.T) {
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
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(nil, assert.AnError).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("payment_collected no top up error", func(t *testing.T) {
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
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
			Error:            "top up not found",
		}

		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(nil, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

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
		amount := float32(5.234)
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, top1).Return(nil, assert.AnError).Once()
		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("syncing top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("success payment_activated", func(t *testing.T) {
		eventType := "payment_activated"
		amount := float32(5.234)
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, top1).Return(&top1, nil).Once()
		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success payment_declined", func(t *testing.T) {
		amount := float32(5.234)
		eventType := "payment_declined"
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, top1).Return(&top1, nil).Once()
		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success payment_cancelled", func(t *testing.T) {
		amount := float32(5.234)
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeInfo,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, top1).Return(&top1, nil).Once()
		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
	})

	t.Run("success force cancel on payment_cancelled sync error", func(t *testing.T) {
		amount := float32(5.234)
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
		top1 := TopUp{
			ID:               paymentID,
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("SyncTopUp", ledgerCtx, top1).Return(nil, assert.AnError).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		lm.On("GetTopUps", ledgerCtx, lcoid).Return([]TopUp{top1}, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, top1).Return(nil).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("syncing top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error force cancel on payment_cancelled sync error payment id not equals top up", func(t *testing.T) {
		amount := float32(5.234)
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
		top1 := TopUp{
			ID:               "id1",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("SyncTopUp", ledgerCtx, top1).Return(nil, assert.AnError).Once()
		lm.On("GetTopUps", ledgerCtx, lcoid).Return([]TopUp{top1}, nil).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("syncing top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error force cancel on payment_cancelled sync error", func(t *testing.T) {
		amount := float32(5.234)
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
		top1 := TopUp{
			ID:               paymentID,
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		top2 := TopUp{
			ID:               "abc",
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}
		sc, _ := json.Marshal(req)
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("SyncTopUp", ledgerCtx, top1).Return(nil, assert.AnError).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		lm.On("GetTopUps", ledgerCtx, lcoid).Return([]TopUp{top1, top2}, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, top1).Return(assert.AnError).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("force cancel top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})

	t.Run("error get top up on payment_cancelled sync error", func(t *testing.T) {
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
		levent := events.Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             events.EventTypeError,
			Action:           events.EventActionDPSWebhookPayment,
			Payload:          sc,
		}
		amount := float32(5.234)
		top1 := TopUp{
			ID:               paymentID,
			LCOrganizationID: lcoid,
			Status:           TopUpStatusActive,
			Amount:           amount,
			Type:             TopUpTypeDirect,
			ConfirmationUrl:  "url",
		}

		lm.On("GetTopUpByIDAndOrganizationID", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("SyncTopUp", ledgerCtx, top1).Return(nil, assert.AnError).Once()
		em.On("ToEvent", ledgerCtx, lcoid, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req).Return(levent).Once()
		lm.On("GetTopUps", ledgerCtx, lcoid).Return(nil, assert.AnError).Once()
		em.On("ToError", ledgerCtx, events.ToErrorParams{
			Event: levent,
			Err:   fmt.Errorf("getting top ups: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
	})
}
