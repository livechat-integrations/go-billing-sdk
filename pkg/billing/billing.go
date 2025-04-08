package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

type (
	BillingEventIDCtxKey              struct{}
	BillingLicenseIDCtxKey            struct{}
	BillingOrganizationIDCtxKey       struct{}
	BillingSubscriptionPlanNameCtxKey struct{}
)

type BillingInterface interface {
	DeleteSubscriptionWithCharge(ctx context.Context, lcOrganizationID string, chargeID string) error
	SyncRecurrentCharge(ctx context.Context, lcOrganizationID string, id string) error
	CreateSubscription(ctx context.Context, lcOrganizationID string, chargeID string, planName string) error
	GetChargesByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Charge, error)
	GetActiveSubscriptionsByOrganizationID(ctx context.Context, lcOrganizationID string) ([]Subscription, error)
}

type Service struct {
	idProvider   common.IdProviderInterface
	billingAPI   livechat.ApiInterface
	eventService common.EventService
	storage      Storage
	plans        Plans
	returnURL    string
	masterOrgID  string
}

func NewService(eventService common.EventService, idProvider common.IdProviderInterface, httpClient *http.Client, livechatEnvironment string, tokenFn livechat.TokenFn, storage Storage, plans Plans, returnUrl, masterOrgID string) *Service {
	a := &livechat.Api{
		HttpClient: httpClient,
		ApiBaseURL: common.EnvURL(livechat.BillingAPIBaseURL, livechatEnvironment),
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
	event := s.eventService.ToEvent(ctx, lcOrganizationID, common.EventActionCreateCharge, common.EventTypeInfo, map[string]interface{}{"name": name, "price": price})
	lcCharge, err := s.billingAPI.CreateRecurrentCharge(ctx, livechat.CreateRecurrentChargeParams{
		Name:      name,
		ReturnURL: s.returnURL,
		Price:     price,
		Test:      s.masterOrgID == lcOrganizationID,
		TrialDays: 0,
		Months:    1,
	})

	if err != nil {
		event.Type = common.EventTypeError
		return "", s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create recurrent charge via lc: %w", err),
		})
	}

	if lcCharge == nil {
		event.Type = common.EventTypeError
		return "", s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create recurrent charge via lc: charge is nil"),
		})
	}

	rawCharge, _ := json.Marshal(lcCharge)
	charge := Charge{
		LCOrganizationID: lcOrganizationID,
		ID:               lcCharge.ID,
		Type:             ChargeTypeRecurring,
		Payload:          rawCharge,
	}
	if err = s.storage.CreateCharge(ctx, charge); err != nil {
		event.Type = common.EventTypeError
		return "", s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create charge in database: %w", err),
		})
	}

	event.SetPayload(charge)
	_ = s.eventService.CreateEvent(ctx, event)

	return lcCharge.ID, nil
}

func (s *Service) SyncRecurrentCharge(ctx context.Context, lcOrganizationID string, id string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, common.EventActionSyncRecurrentCharge, common.EventTypeInfo, map[string]interface{}{"id": id})
	charge, err := s.storage.GetCharge(ctx, id)
	if err != nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get charge: %w", err),
		})
	}

	if charge == nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("charge not found"),
		})
	}

	lcCharge, err := s.billingAPI.GetRecurrentCharge(ctx, id)
	if err != nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get recurrent charge: %w", err),
		})
	}

	if err = s.storage.UpdateChargePayload(ctx, id, lcCharge.BaseCharge); err != nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
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

	res := []Subscription{}
	for _, sub := range subs {
		if sub.IsActive() {
			res = append(res, sub)
		}
	}
	return res, nil
}

func (s *Service) CreateSubscription(ctx context.Context, lcOrganizationID string, chargeID string, planName string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, common.EventActionCreateSubscription, common.EventTypeInfo, map[string]interface{}{"planName": planName, "chargeID": chargeID})
	plan := s.plans.GetPlan(planName)
	if plan == nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("plan not found"),
		})
	}

	charge, err := s.storage.GetCharge(ctx, chargeID)
	if err != nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to get charge by organization id: %w", err),
		})
	}

	if charge == nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
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
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create subscription in database: %w", err),
		})
	}

	event.SetPayload(charge)
	_ = s.eventService.CreateEvent(ctx, event)

	return nil
}

func (s *Service) DeleteSubscriptionWithCharge(ctx context.Context, lcOrganizationID string, chargeID string) error {
	event := s.eventService.ToEvent(ctx, lcOrganizationID, common.EventActionDeleteSubscriptionWithCharge, common.EventTypeInfo, map[string]interface{}{"chargeID": chargeID})
	if err := s.storage.DeleteSubscriptionByChargeID(ctx, lcOrganizationID, chargeID); err != nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to delete subscription: %w", err),
		})
	}

	if err := s.storage.DeleteCharge(ctx, chargeID); err != nil {
		event.Type = common.EventTypeError
		return s.eventService.ToError(ctx, common.ToErrorParams{
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
