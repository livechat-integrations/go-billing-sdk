package billing

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
	billing      BillingInterface
	idProvider   events.IdProviderInterface
	eventService events.EventService
}

type HandlerInterface interface {
	HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error
}

func NewHandler(eventService events.EventService, billing BillingInterface, idProvider events.IdProviderInterface) *Handler {
	return &Handler{
		billing:      billing,
		idProvider:   idProvider,
		eventService: eventService,
	}
}

func (h *Handler) HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error {
	ctx = context.WithValue(ctx, BillingEventIDCtxKey{}, h.idProvider.GenerateId())
	ctx = context.WithValue(ctx, BillingOrganizationIDCtxKey{}, req.LCOrganizationID)
	ctx = context.WithValue(ctx, BillingLicenseIDCtxKey{}, req.License)
	chargeID, exists := req.Payload["paymentID"].(string)
	if !exists {
		return nil
	}

	event := h.eventService.ToEvent(ctx, req.LCOrganizationID, events.EventActionUnknown, events.EventTypeInfo, req)

	switch req.Event {
	case "application_uninstalled":
		event.Action = events.EventActionDPSWebhookApplicationUninstalled
		charges, err := h.billing.GetChargesByOrganizationID(ctx, req.LCOrganizationID)
		if err != nil {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}

		for _, charge := range charges {
			if err := h.billing.DeleteSubscriptionWithCharge(ctx, req.LCOrganizationID, charge.ID); err != nil {
				event.Type = events.EventTypeError
				return h.eventService.ToError(ctx, events.ToErrorParams{
					Event: event,
					Err:   fmt.Errorf("delete subscription with charge: %w", err),
				})
			}
		}
	case "payment_cancelled":
		event.Action = events.EventActionDPSWebhookPayment
		if err := h.billing.DeleteSubscriptionWithCharge(ctx, req.LCOrganizationID, chargeID); err != nil {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("delete subscription with charge: %w", err),
			})
		}
	case "payment_activated", "payment_collected":
		event.Action = events.EventActionDPSWebhookPayment

		if err := h.billing.SyncRecurrentCharge(ctx, req.LCOrganizationID, chargeID); err != nil {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("sync recurrent charge: %w", err),
			})
		}

		if req.Event != "payment_activated" {
			break
		}

		planName, ok := ctx.Value(BillingSubscriptionPlanNameCtxKey{}).(string)
		if !ok || planName == "" {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("no plan name found in context"),
			})
		}

		if err := h.billing.CreateSubscription(ctx, req.LCOrganizationID, chargeID, planName); err != nil {
			event.Type = events.EventTypeError
			return h.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("create subscription: %w", err),
			})
		}
	}

	_ = h.eventService.CreateEvent(ctx, event)

	return nil
}
