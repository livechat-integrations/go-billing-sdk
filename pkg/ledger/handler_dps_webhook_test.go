package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var lm = new(ledgerMock)

var h = Handler{
	ledger:     lm,
	idProvider: xm,
}

var xid = "2341"
var ledgerCtx = context.WithValue(context.Background(), ledgerEventIDCtxKey{}, xid)

type ledgerMock struct {
	mock.Mock
}

func (l *ledgerMock) CreateCharge(ctx context.Context, params CreateChargeParams) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) GetTopUps(ctx context.Context, organizationID string) ([]TopUp, error) {
	//TODO implement me
	panic("implement me")
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

func (l *ledgerMock) ForceCancelTopUp(ctx context.Context, organizationID string, ID string) error {
	args := l.Called(ctx, organizationID, ID)
	return args.Error(0)
}

func (l *ledgerMock) CancelCharge(ctx context.Context, organizationID string, ID string) error {
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

func (l *ledgerMock) ToError(ctx context.Context, params ToErrorParams) error {
	args := l.Called(ctx, params)
	return args.Error(0)
}

func (l *ledgerMock) ToEvent(id string, organizationID string, action EventAction, eventType EventType, payload any) Event {
	args := l.Called(id, organizationID, action, eventType, payload)
	return args.Get(0).(Event)
}

func (l *ledgerMock) GetUniqueID() string {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) CreateEvent(ctx context.Context, event Event) error {
	args := l.Called(ctx, event)
	return args.Error(0)
}

func (l *ledgerMock) UpdateTopUpStatus(ctx context.Context, organizationID string, ID string, status TopUpStatus) error {
	//TODO implement me
	panic("implement me")
}

func (l *ledgerMock) SyncTopUp(ctx context.Context, organizationID string, ID string) (*TopUp, error) {
	args := l.Called(ctx, organizationID, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopUp), args.Error(1)
}

func TestNewHandler(t *testing.T) {
	t.Run("NewHandler", func(t *testing.T) {
		newService := NewHandler(&ledgerMock{}, &xIdMock{})

		assert.NotNil(t, newService)
		assertExpectations(t)
	})
}

func TestService_HandleDPSWebhook(t *testing.T) {
	t.Run("success application_uninstalled", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
		}
		sc, _ := json.Marshal(req)
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return([]TopUp{top1, top2}, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, lcoid, top1.ID).Return(nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, lcoid, top2.ID).Return(nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookApplicationUninstalled, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("success application_uninstalled no top ups", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return([]TopUp{}, nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookApplicationUninstalled, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("application_uninstalled get top ups error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return(nil, assert.AnError).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookApplicationUninstalled, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("application_uninstalled force cancel error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "application_uninstalled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		someDate2, _ := time.Parse(time.DateTime, "2025-06-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
			CurrentToppedUpAt: someDate,
			NextTopUpAt:       someDate2,
		}
		sc, _ := json.Marshal(req)
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookApplicationUninstalled,
			Payload:          sc,
		}

		lm.On("GetTopUpsByOrganizationIDAndStatus", ledgerCtx, lcoid, TopUpStatusActive).Return([]TopUp{top1, top2}, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, lcoid, top1.ID).Return(nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, lcoid, top2.ID).Return(assert.AnError).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookApplicationUninstalled, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   assert.AnError,
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("success payment_collected", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_collected"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("payment_collected wrong payment id", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_collected"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := float32(234.3)
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   fmt.Errorf("payment id field not found in payload"),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("payment_collected sync error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		eventType := "payment_collected"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(nil, assert.AnError).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   fmt.Errorf("syncing top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("success payment_activated", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_activated"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("success payment_declined", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_declined"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("success payment_cancelled", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_cancelled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeInfo,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(&top1, nil).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.Nil(t, err)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("success force cancel on payment_cancelled sync error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_cancelled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(nil, assert.AnError).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("GetTopUpByID", ledgerCtx, paymentID).Return(&top1, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, lcoid, top1.ID).Return(nil).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   fmt.Errorf("syncing top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("error force cancel on payment_cancelled sync error", func(t *testing.T) {
		amount := float32(5.234)
		lcoid := "lcOrganizationID"
		eventType := "payment_cancelled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(nil, assert.AnError).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("GetTopUpByID", ledgerCtx, paymentID).Return(&top1, nil).Once()
		lm.On("ForceCancelTopUp", ledgerCtx, lcoid, top1.ID).Return(assert.AnError).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   fmt.Errorf("force cancell top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})

	t.Run("error get top up on payment_cancelled sync error", func(t *testing.T) {
		lcoid := "lcOrganizationID"
		eventType := "payment_cancelled"
		someDate, _ := time.Parse(time.DateTime, "2025-03-14 12:31:56")
		paymentID := "x1c2v3"
		userID := "s98f"
		call := xm.On("GenerateId").Return(xid, nil)

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
		levent := Event{
			ID:               xid,
			LCOrganizationID: lcoid,
			Type:             EventTypeError,
			Action:           EventActionDPSWebhookPayment,
			Payload:          sc,
		}

		lm.On("SyncTopUp", ledgerCtx, lcoid, paymentID).Return(nil, assert.AnError).Once()
		lm.On("ToEvent", xid, lcoid, EventActionDPSWebhookPayment, EventTypeInfo, req).Return(levent).Once()
		lm.On("CreateEvent", ledgerCtx, levent).Return(nil).Once()
		lm.On("GetTopUpByID", ledgerCtx, paymentID).Return(nil, assert.AnError).Once()
		lm.On("ToError", ledgerCtx, ToErrorParams{
			event: levent,
			err:   fmt.Errorf("getting top up: %w", assert.AnError),
		}).Return(assert.AnError).Once()

		err := h.HandleDPSWebhook(context.Background(), req)

		assert.ErrorIs(t, err, assert.AnError)

		assertExpectations(t)
		call.Unset()
	})
}
