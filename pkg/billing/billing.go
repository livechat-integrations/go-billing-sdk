package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/v2/common"
	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

type (
	EventIDCtxKey              struct{}
	LicenseIDCtxKey            struct{}
	OrganizationIDCtxKey       struct{}
	SubscriptionPlanNameCtxKey struct{}
)

type ServiceInterface interface {
	DeleteSubscriptionWithCharge(ctx context.Context, lcOrganizationID string, chargeID string) error
	DeleteSubscription(ctx context.Context, lcOrganizationID string, subscriptionID string) error
	SyncRecurrentCharge(ctx context.Context, lcOrganizationID string, id string) error
	CreateSubscription(ctx context.Context, lcOrganizationID string, chargeID string, planName string) error
	GetChargesByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Charge, error)
	GetActiveSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error)
	GetSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error)
	SyncCharges(ctx context.Context) error
}

type Service struct {
	idProvider   events.IdProviderInterface
	billingAPI   livechat.ApiInterface
	eventService events.EventService
	storage      Storage
	plans        Plans
	returnURL    string
	masterOrgID  string
}

func NewService(eventService events.EventService, idProvider events.IdProviderInterface, httpClient *http.Client, livechatEnvironment string, tokenFn common.TokenFn, storage Storage, plans Plans, returnUrl, masterOrgID string) *Service {
	a := &livechat.Api{
		HttpClient: httpClient,
		ApiBaseURL: events.EnvURL(livechat.BillingAPIBaseURL, livechatEnvironment),
		TokenFn:    tokenFn,
	}

	return &Service{
		billingAPI:   a,
		eventService: eventService,
		idProvider:   idProvider,
		storage:      storage,
		plans:        plans,
		returnURL:    returnUrl,
		masterOrgID:  masterOrgID,
	}
}

func (s *Service) CreateRecurrentCharge(ctx context.Context, name string, price int, lcOrganizationID string) (string, error) {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, events.EventActionCreateCharge, events.EventTypeInfo, map[string]interface{}{"name": name, "price": price})
	lcCharge, err := s.billingAPI.CreateRecurrentCharge(ctx, livechat.CreateRecurrentChargeParams{
		Name:      name,
		ReturnURL: s.returnURL,
		Price:     price,
		Test:      s.masterOrgID == lcOrganizationID,
		TrialDays: 0,
		Months:    1,
	})

	if err != nil {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create recurrent charge via lc: %w", err),
		})
	}

	if lcCharge == nil {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create recurrent charge via lc: charge is nil"),
		})
	}

	rawCharge, _ := json.Marshal(lcCharge)
	charge := Charge{
		LCOrganizationID: lcOrganizationID,
		ID:               lcCharge.ID,
		Type:             ChargeTypeRecurring,
		Status:           livechat.RecurrentChargeStatusPending,
		Payload:          rawCharge,
	}
	if err = s.storage.CreateCharge(ctx, charge); err != nil {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create charge in database: %w", err),
		})
	}

	event.SetPayload(charge)
	_ = s.eventService.CreateEvent(ctx, event)

	return lcCharge.ID, nil
}

func (s *Service) SyncRecurrentCharge(ctx context.Context, lcOrganizationID string, id string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, map[string]interface{}{"id": id})
	charge, err := s.storage.GetCharge(ctx, id)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get charge: %w", err),
		})
	}

	if charge == nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("charge not found"),
		})
	}

	lcCharge, err := s.billingAPI.GetRecurrentCharge(ctx, id)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get recurrent charge: %w", err),
		})
	}

	rawCharge, _ := json.Marshal(lcCharge)
	if err = s.storage.UpdateChargePayload(ctx, id, rawCharge); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to update charge payload: %w", err),
		})
	}

	event.SetPayload(lcCharge)
	_ = s.eventService.CreateEvent(ctx, event)

	return nil
}

func (s *Service) GetCharge(ctx context.Context, id string) (*Charge, error) {
	return s.storage.GetCharge(ctx, id)
}

func (s *Service) IsPremium(ctx context.Context, id string) (bool, error) {
	sub, err := s.storage.GetSubscriptionsByOrganizationID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to get charge by installation id: %w", err)
	}

	//TODO: Refactor when we have multiple subscriptions
	return len(sub) > 0 && sub[0].IsActive(), nil
}

func (s *Service) GetActiveSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error) {
	subs, err := s.storage.GetSubscriptionsByOrganizationID(ctx, lcOrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by organization id: %w", err)
	}

	var res []Subscription
	for _, sub := range subs {
		if sub.IsActive() {
			res = append(res, sub)
		}
	}
	return res, nil
}

func (s *Service) GetSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error) {
	subs, err := s.storage.GetSubscriptionsByOrganizationID(ctx, lcOrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions by organization id: %w", err)
	}

	return subs, nil
}

func (s *Service) CreateSubscription(ctx context.Context, lcOrganizationID string, chargeID string, planName string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, events.EventActionCreateSubscription, events.EventTypeInfo, map[string]interface{}{"planName": planName, "chargeID": chargeID})

	dbSubscriptions, err := s.storage.GetSubscriptionsByOrganizationID(ctx, lcOrganizationID)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get subscriptions by organization id: %w", err),
		})
	}

	for _, sub := range dbSubscriptions {
		if sub.Charge.ID == chargeID {
			event.SetPayload(map[string]interface{}{"planName": planName, "chargeID": chargeID, "result": "subscription already exists"})
			_ = s.eventService.CreateEvent(ctx, event)
			return nil
		}
	}

	plan := s.plans.GetPlan(planName)
	if plan == nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("plan not found"),
		})
	}

	charge, err := s.storage.GetCharge(ctx, chargeID)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get charge by organization id: %w", err),
		})
	}

	if charge == nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("charge not found"),
		})
	}

	if err = s.storage.CreateSubscription(ctx, Subscription{
		ID:               s.idProvider.GenerateId(),
		Charge:           charge,
		LCOrganizationID: lcOrganizationID,
		PlanName:         planName,
	}); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create subscription in database: %w", err),
		})
	}

	event.SetPayload(charge)
	_ = s.eventService.CreateEvent(ctx, event)

	return nil
}

func (s *Service) DeleteSubscriptionWithCharge(ctx context.Context, lcOrganizationID string, chargeID string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, events.EventActionDeleteSubscriptionWithCharge, events.EventTypeInfo, map[string]interface{}{"chargeID": chargeID})
	if err := s.storage.DeleteSubscriptionByChargeID(ctx, lcOrganizationID, chargeID); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete subscription: %w", err),
		})
	}

	if err := s.storage.UpdateChargeStatus(ctx, chargeID, livechat.RecurrentChargeStatusCancelled); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete subscription: %w", err),
		})
	}

	if err := s.storage.DeleteCharge(ctx, chargeID); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete charge: %w", err),
		})
	}

	_ = s.eventService.CreateEvent(ctx, event)

	return nil
}

func (s *Service) GetChargesByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Charge, error) {
	rows, err := s.storage.GetChargesByOrganizationID(ctx, lcOrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get charges by organization id: %w", err)
	}

	return rows, nil
}

func (s *Service) CancelRecurrentCharge(ctx context.Context, chargeID string) error {
	recCharge, _ := s.billingAPI.CancelRecurrentCharge(ctx, chargeID)

	rawCharge, _ := json.Marshal(recCharge)
	if err := s.storage.UpdateChargePayload(ctx, chargeID, rawCharge); err != nil {
		return fmt.Errorf("failed to update charge payload: %w", err)
	}

	return nil
}

func (s *Service) SyncCharges(ctx context.Context) error {
	charges, err := s.storage.GetChargesByStatuses(ctx, GetSyncValidStatuses())
	if err != nil {
		return fmt.Errorf("failed to get charges by statuses: %w", err)
	}

	for _, charge := range charges {
		organizationCtx := context.WithValue(ctx, OrganizationIDCtxKey{}, charge.LCOrganizationID)
		organizationCtx = context.WithValue(organizationCtx, EventIDCtxKey{}, s.idProvider.GenerateId())

		var recCharge livechat.RecurrentCharge
		_ = json.Unmarshal(charge.Payload, &recCharge)
		if recCharge.Status == livechat.RecurrentChargeStatusActive && recCharge.NextChargeAt.Before(time.Now()) {
			continue
		}

		if recCharge.Status == livechat.RecurrentChargeStatusPending && recCharge.CreatedAt.AddDate(0, 1, 0).Before(time.Now()) {
			if err = s.cancelChange(ctx, charge); err != nil {
				return err
			}

			continue
		}

		event := s.eventService.ToEvent(organizationCtx, charge.LCOrganizationID, events.EventActionSyncRecurrentCharge, events.EventTypeInfo, map[string]interface{}{"id": charge.ID})
		lcCharge, err := s.billingAPI.GetRecurrentCharge(organizationCtx, charge.ID)
		if err != nil {
			event.Type = events.EventTypeError
			return s.eventService.ToError(organizationCtx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("failed to get recurrent charge: %w", err),
			})
		}

		switch lcCharge.Status {
		case livechat.RecurrentChargeStatusAccepted,
			livechat.RecurrentChargeStatusFrozen:
			lcCharge, err = s.billingAPI.ActivateRecurrentCharge(organizationCtx, charge.ID)
			if err != nil {
				event.Type = events.EventTypeError
				return s.eventService.ToError(organizationCtx, events.ToErrorParams{
					Event: event,
					Err:   fmt.Errorf("failed to activate charge: %w", err),
				})
			}
		}

		rawCharge, _ := json.Marshal(recCharge)
		if err = s.storage.UpdateChargePayload(organizationCtx, charge.ID, rawCharge); err != nil {
			event.Type = events.EventTypeError
			return s.eventService.ToError(organizationCtx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("failed to update charge payload: %w", err),
			})
		}

		if err = s.storage.UpdateChargeStatus(organizationCtx, charge.ID, charge.Status); err != nil {
			event.Type = events.EventTypeError
			return s.eventService.ToError(organizationCtx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("failed to update charge: %w", err),
			})
		}

		event.SetPayload(lcCharge)
		_ = s.eventService.CreateEvent(organizationCtx, event)
	}

	return nil
}

func (s *Service) DeleteSubscription(ctx context.Context, lcOrganizationID, subscriptionID string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, events.EventActionDeleteSubscription, events.EventTypeInfo, map[string]interface{}{"subscriptionID": subscriptionID})
	subs, err := s.storage.GetSubscriptionsByOrganizationID(ctx, lcOrganizationID)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get subscriptions: %w", err),
		})
	}

	var sub *Subscription
	for _, subDB := range subs {
		if subDB.ID == subscriptionID {
			sub = &subDB
		}
	}

	if sub == nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("subscription %s not found", subscriptionID),
		})
	}

	if err = s.storage.DeleteSubscription(ctx, lcOrganizationID, subscriptionID); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete subscription: %w", err),
		})
	}

	if err = s.storage.UpdateChargeStatus(ctx, sub.Charge.ID, livechat.RecurrentChargeStatusCancelled); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete subscription: %w", err),
		})
	}

	if err = s.storage.DeleteCharge(ctx, sub.Charge.ID); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete charge: %w", err),
		})
	}

	_ = s.eventService.CreateEvent(ctx, event)

	return nil
}

func (s *Service) cancelChange(ctx context.Context, charge Charge) error {
	event := s.eventService.ToEvent(ctx, charge.LCOrganizationID, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": charge.ID})
	cancelledCharge, err := s.billingAPI.CancelRecurrentCharge(ctx, charge.ID)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to cancel charge: %w", err),
		})
	}

	rawCharge, _ := json.Marshal(cancelledCharge)
	if err = s.storage.UpdateChargePayload(ctx, charge.ID, rawCharge); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to cancel charge: %w", err),
		})
	}

	if err = s.storage.UpdateChargeStatus(ctx, charge.ID, livechat.RecurrentChargeStatusCancelled); err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to update charge: %w", err),
		})
	}

	_ = s.eventService.CreateEvent(ctx, event)
	return nil
}
