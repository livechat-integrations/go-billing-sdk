package common

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
	EventActionDeleteSubscriptionWithCharge     EventAction = "delete_subscription_with_charge"
	EventActionSyncRecurrentCharge              EventAction = "sync_recurrent_charge"
	EventActionCreateSubscription               EventAction = "create_subscription"
	EventActionCreateCharge                     EventAction = "create_charge"
	EventActionCreateTopUp                      EventAction = "create_top_up"
	EventActionCancelTopUp                      EventAction = "cancel_top_up"
	EventActionCancelCharge                     EventAction = "cancel_charge_event"
	EventActionForceCancelCharge                EventAction = "force_cancel_charge_event"
	EventActionDPSWebhookApplicationUninstalled EventAction = "dps_webhook_event_application_uninstalled"
	EventActionDPSWebhookPayment                EventAction = "dps_webhook_event_payment"
	EventActionSyncTopUp                        EventAction = "sync_top_up_event"
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
