package ledger

import (
	"context"
	"fmt"
	"time"
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
	ledger LedgerInterface
}

func NewHandler(ledger LedgerInterface) *Handler {
	return &Handler{
		ledger: ledger,
	}
}

func (h *Handler) HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error {
	event := h.ledger.ToEvent(h.ledger.GetUniqueID(), req.LCOrganizationID, EventActionDPSWebhook, EventTypeInfo, req)
	switch req.Event {
	case "application_uninstalled":
		topUps, err := h.ledger.GetTopUpsByOrganizationIDAndStatus(ctx, req.LCOrganizationID, TopUpStatusActive)
		if err != nil {
			event.Type = EventTypeError
			return h.ledger.ToError(ctx, ToErrorParams{
				event: event,
				err:   err,
			})
		}
		for _, t := range topUps {
			if err := h.ledger.ForceCancelTopUp(ctx, t.LCOrganizationID, t.ID); err != nil {
				return err
			}
		}
	case "payment_collected", "payment_activated", "payment_cancelled", "payment_declined":
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
				if t, err := h.ledger.GetTopUpByID(ctx, paymentID); err == nil {
					if err := h.ledger.ForceCancelTopUp(ctx, req.LCOrganizationID, t.ID); err != nil {
						return err
					}
				} else {
					event.Type = EventTypeError
					return h.ledger.ToError(ctx, ToErrorParams{
						event: event,
						err:   fmt.Errorf("getting charge by lc_id: %w", err),
					})
				}

			}
			event.Type = EventTypeError
			return h.ledger.ToError(ctx, ToErrorParams{
				event: event,
				err:   fmt.Errorf("syncing charge: %w", err),
			})
		}
	}

	_ = h.ledger.CreateEvent(ctx, event)

	return nil
}
