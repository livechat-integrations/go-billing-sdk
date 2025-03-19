package ledger

import (
	"encoding/json"
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
	EventActionCreateTopUp                      EventAction = "create_top_up"
	EventActionCancelTopUp                      EventAction = "cancel_top_up"
	EventActionUpdateTopUpStatus                EventAction = "update_top_up_status"
	EventActionCancelCharge                     EventAction = "cancel_charge_event"
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
