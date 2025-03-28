package ledger

import (
	"context"
	"fmt"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
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
	ledger     LedgerInterface
	idProvider common.IdProviderInterface
}

type HandlerInterface interface {
	HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error
}

func NewHandler(ledger LedgerInterface, idProvider common.IdProviderInterface) *Handler {
	return &Handler{
		ledger:     ledger,
		idProvider: idProvider,
	}
}

func (h *Handler) HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error {
	ctx = context.WithValue(ctx, LedgerEventIDCtxKey{}, h.idProvider.GenerateId())
	ctx = context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, req.LCOrganizationID)
	switch req.Event {
	case "application_uninstalled":
		event := h.ledger.ToEvent(ctx, req.LCOrganizationID, EventActionDPSWebhookApplicationUninstalled, EventTypeInfo, req)
		topUps, err := h.ledger.GetTopUpsByOrganizationIDAndStatus(ctx, req.LCOrganizationID, TopUpStatusActive)
		if err != nil {
			event.Type = EventTypeError
			return h.ledger.ToError(ctx, ToErrorParams{
				event: event,
				err:   err,
			})
		}
		for _, t := range topUps {
			if err := h.ledger.ForceCancelTopUp(ctx, t); err != nil {
				event.Type = EventTypeError
				return h.ledger.ToError(ctx, ToErrorParams{
					event: event,
					err:   err,
				})
			}
		}
		_ = h.ledger.CreateEvent(ctx, event)
	case "payment_collected", "payment_cancelled", "payment_declined":
		event := h.ledger.ToEvent(ctx, req.LCOrganizationID, EventActionDPSWebhookPayment, EventTypeInfo, req)
		paymentID, ok := req.Payload["paymentID"].(string)
		if !ok {
			event.Type = EventTypeError
			return h.ledger.ToError(ctx, ToErrorParams{
				event: event,
				err:   fmt.Errorf("payment id field not found in payload"),
			})
		}
		_, err := h.ledger.SyncTopUp(ctx, req.LCOrganizationID, paymentID)
		if err != nil {
			if req.Event == "payment_cancelled" {
				if tps, err := h.ledger.GetTopUps(ctx, req.LCOrganizationID); err == nil {
					for _, t := range tps {
						if paymentID != t.ID {
							continue
						}
						if err := h.ledger.ForceCancelTopUp(ctx, t); err != nil {
							event.Type = EventTypeError
							return h.ledger.ToError(ctx, ToErrorParams{
								event: event,
								err:   fmt.Errorf("force cancell top up: %w", err),
							})
						}
					}
				} else {
					event.Type = EventTypeError
					return h.ledger.ToError(ctx, ToErrorParams{
						event: event,
						err:   fmt.Errorf("getting top up: %w", err),
					})
				}

			}
			event.Type = EventTypeError
			return h.ledger.ToError(ctx, ToErrorParams{
				event: event,
				err:   fmt.Errorf("syncing top up: %w", err),
			})
		}

		_ = h.ledger.CreateEvent(ctx, event)
	}

	return nil
}
