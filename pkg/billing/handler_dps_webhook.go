package billing

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
	billing      BillingInterface
	idProvider   common.IdProviderInterface
	eventService common.EventService
}

type HandlerInterface interface {
	HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error
}

func NewHandler(eventService common.EventService, billing BillingInterface, idProvider common.IdProviderInterface) *Handler {
	return &Handler{
		billing:      billing,
		idProvider:   idProvider,
		eventService: eventService,
	}
}

func (h *Handler) HandleDPSWebhook(ctx context.Context, req DPSWebhookRequest) error {
	ctx = context.WithValue(ctx, BillingEventIDCtxKey{}, h.idProvider.GenerateId())
	ctx = context.WithValue(ctx, BillingOrganizationIDCtxKey{}, req.LCOrganizationID)
	chargeID, exists := req.Payload["paymentID"].(string)
	if !exists {
		return nil
	}

	switch req.Event {
	case "application_uninstalled":
		event := h.eventService.ToEvent(ctx, req.LCOrganizationID, common.EventActionDPSWebhookApplicationUninstalled, common.EventTypeInfo, req)
		charges, err := h.billing.GetChargesByOrganizationID(ctx, req.LCOrganizationID)
		if err != nil {
			event.Type = common.EventTypeError
			return h.eventService.ToError(ctx, common.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}

		for _, charge := range charges {
			if err := h.billing.DeleteSubscriptionWithCharge(ctx, req.LCOrganizationID, charge.ID); err != nil {
				event.Type = common.EventTypeError
				return h.eventService.ToError(ctx, common.ToErrorParams{
					Event: event,
					Err:   fmt.Errorf("delete subscription with charge: %w", err),
				})
			}
		}
		_ = h.eventService.CreateEvent(ctx, event)
	case "payment_collected", "payment_activated":
		event := h.eventService.ToEvent(ctx, req.LCOrganizationID, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req)
		if err := h.billing.SyncRecurrentCharge(ctx, req.LCOrganizationID, chargeID); err != nil {
			event.Type = common.EventTypeError
			return h.eventService.ToError(ctx, common.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("sync recurrent charge: %w", err),
			})
		}

		planName, ok := ctx.Value(BillingSubscriptionPlanNameCtxKey{}).(string)
		if !ok || planName == "" {
			event.Type = common.EventTypeError
			return h.eventService.ToError(ctx, common.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("no plan name found in context"),
			})
		}

		if err := h.billing.CreateSubscription(ctx, req.LCOrganizationID, chargeID, planName); err != nil {
			event.Type = common.EventTypeError
			return h.eventService.ToError(ctx, common.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("create subscription: %w", err),
			})
		}
		_ = h.eventService.CreateEvent(ctx, event)
	case "payment_cancelled":
		event := h.eventService.ToEvent(ctx, req.LCOrganizationID, common.EventActionDPSWebhookPayment, common.EventTypeInfo, req)

		if err := h.billing.DeleteSubscriptionWithCharge(ctx, req.LCOrganizationID, chargeID); err != nil {
			event.Type = common.EventTypeError
			return h.eventService.ToError(ctx, common.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("delete subscription with charge: %w", err),
			})
		}
		_ = h.eventService.CreateEvent(ctx, event)
	}

	return nil
}
