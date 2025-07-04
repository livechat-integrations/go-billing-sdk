package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type EventType string

const (
	EventTypeInfo  EventType = "info"
	EventTypeError EventType = "error"
)

type EventAction string

const (
	EventActionCreateCharge                     EventAction = "create_charge"
	EventActionDeleteSubscriptionWithCharge     EventAction = "delete_subscription_with_charge"
	EventActionDeleteSubscription               EventAction = "delete_subscription"
	EventActionSyncRecurrentCharge              EventAction = "sync_recurrent_charge"
	EventActionCreateSubscription               EventAction = "create_subscription"
	EventActionCreateOperation                  EventAction = "create_operation"
	EventActionTopUp                            EventAction = "top_up"
	EventActionCreateTopUp                      EventAction = "create_top_up"
	EventActionCancelRecurrentTopUp             EventAction = "cancel_recurrent_top_up"
	EventActionForceCancelCharge                EventAction = "force_cancel_charge_event"
	EventActionDPSWebhookApplicationUninstalled EventAction = "dps_webhook_event_application_uninstalled"
	EventActionDPSWebhookPayment                EventAction = "dps_webhook_event_payment"
	EventActionSyncTopUp                        EventAction = "sync_top_up_event"
	EventActionActivateCharge                   EventAction = "activate_charge"
	EventActionAddVoucherFunds                  EventAction = "add_voucher_funds"
	EventActionUnknown                          EventAction = "unknown"
)

type Event struct {
	ID               string
	LCOrganizationID string
	Type             EventType
	Action           EventAction
	Payload          json.RawMessage
	Error            string
	CreatedAt        time.Time
}

func (e *Event) SetPayload(payload any) {
	jp, err := json.Marshal(payload)
	if err == nil {
		e.Payload = jp
	}
}

type Storage interface {
	CreateEvent(context.Context, Event) error
}

type EventService interface {
	CreateEvent(ctx context.Context, event Event) error
	ToError(ctx context.Context, params ToErrorParams) error
	ToEvent(ctx context.Context, organizationID string, action EventAction, eventType EventType, payload any) Event
}

type Service struct {
	storage       Storage
	idProvider    IdProviderInterface
	eventIdCtxKey interface{}
}

func NewService(storage Storage, idProvider IdProviderInterface, eventIdCtxKey interface{}) *Service {
	return &Service{
		storage:       storage,
		idProvider:    idProvider,
		eventIdCtxKey: eventIdCtxKey,
	}
}

func (s *Service) CreateEvent(ctx context.Context, event Event) error {
	err := s.storage.CreateEvent(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

type ToErrorParams struct {
	Event Event
	Err   error
}

func (s *Service) ToError(ctx context.Context, params ToErrorParams) error {
	params.Event.Error = params.Err.Error()
	_ = s.CreateEvent(ctx, params.Event)
	return fmt.Errorf("%s: %w", params.Event.ID, params.Err)
}

func (s *Service) ToEvent(ctx context.Context, organizationID string, action EventAction, eventType EventType, payload any) Event {
	id, ok := ctx.Value(s.eventIdCtxKey).(string)
	if !ok {
		id = s.idProvider.GenerateId()
	}

	event := Event{
		ID:               id,
		LCOrganizationID: organizationID,
		Type:             eventType,
		Action:           action,
		CreatedAt:        time.Time{},
	}
	jp, err := json.Marshal(payload)
	if err == nil {
		event.Payload = jp
	}

	return event
}
