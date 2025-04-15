package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/pkg/events"
)

type DPSWebhookRequest struct {
	ApplicationID    string                 `json:"applicationID"`
	ApplicationName  string                 `json:"applicationName"`
	ClientID         string                 `json:"clientID"`
	Date             time.Time              `json:"date"`
	Event            string                 `json:"event"`
	License          int32                  `json:"licenseID"`
	LCOrganizationID string                 `json:"organizationID"`
	Payload          map[string]interface{} `json:"payload"`
	UserID           string                 `json:"userID"`
}

type Handler struct {
	ledger       LedgerInterface
	idProvider   events.IdProviderInterface
	eventService events.EventService
}

type HandlerInterface interface {
	HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error
}

func NewHandler(eventService events.EventService, ledger LedgerInterface, idProvider events.IdProviderInterface) *Handler {
	return &Handler{
		ledger:       ledger,
		idProvider:   idProvider,
		eventService: eventService,
	}
}

func (h *Handler) HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error {
	ctx = context.WithValue(ctx, LedgerEventIDCtxKey{}, h.idProvider.GenerateId())
	ctx = context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, req.LCOrganizationID)
	switch req.Event {
	case "application_uninstalled":
		event := h.eventService.ToEvent(ctx, req.LCOrganizationID, events.EventActionDPSWebhookApplicationUninstalled, events.EventTypeInfo, req)
		topUps, err := h.ledger.GetTopUpsByOrganizationIDAndStatus(ctx, req.LCOrganizationID, TopUpStatusActive)
		if err != nil {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}
		for _, t := range topUps {
			if err := h.ledger.ForceCancelTopUp(ctx, t); err != nil {
				event.Type = events.EventTypeError
				return h.eventService.ToError(ctx, events.ToErrorParams{
					Event: event,
					Err:   err,
				})
			}
		}
		_ = h.eventService.CreateEvent(ctx, event)
	case "payment_collected", "payment_activated", "payment_cancelled", "payment_declined":
		event := h.eventService.ToEvent(ctx, req.LCOrganizationID, events.EventActionDPSWebhookPayment, events.EventTypeInfo, req)
		paymentID, ok := req.Payload["paymentID"].(string)
		if !ok {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("payment id field not found in payload"),
			})
		}
		dbTopUp, err := h.ledger.GetTopUpByIDAndOrganizationID(ctx, req.LCOrganizationID, paymentID)
		if err != nil {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}
		if dbTopUp == nil {
			event.Error = "top up not found"
			_ = h.eventService.CreateEvent(ctx, event)
			return nil
		}
		topUp, err := h.ledger.SyncTopUp(ctx, *dbTopUp)
		if err != nil {
			if req.Event == "payment_cancelled" {
				if tps, err := h.ledger.GetTopUps(ctx, req.LCOrganizationID); err == nil {
					for _, t := range tps {
						if paymentID != t.ID {
							continue
						}
						if err := h.ledger.ForceCancelTopUp(ctx, t); err != nil {
							event.Type = events.EventTypeError
							return h.eventService.ToError(ctx, events.ToErrorParams{
								Event: event,
								Err:   fmt.Errorf("force cancell top up: %w", err),
							})
						}
					}
				} else {
					event.Type = events.EventTypeError
					return h.eventService.ToError(ctx, events.ToErrorParams{
						Event: event,
						Err:   fmt.Errorf("getting top up: %w", err),
					})
				}

			}
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("syncing top up: %w", err),
			})
		}

		if req.Event == "payment_collected" {
			_, err2 := h.ledger.TopUp(ctx, *topUp)
			if err2 != nil {
				event.Type = events.EventTypeError
				return h.eventService.ToError(ctx, events.ToErrorParams{
					Event: event,
					Err:   fmt.Errorf("top up: %w", err2),
				})
			}
		}

		_ = h.eventService.CreateEvent(ctx, event)
	}

	return nil
}
