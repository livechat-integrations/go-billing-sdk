package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"net/http"
)

type Service struct {
	billingAPI  apiInterface
	storage     Storage
	plans       Plans
	returnURL   string
	masterOrgID string
}

func NewService(httpClient *http.Client, livechatEnvironment string, tokenFn TokenFn, storage Storage, plans Plans, returnUrl, masterOrgID string) *Service {
	a := &api{
		httpClient: httpClient,
		apiBaseURL: EnvURL(billingAPIBaseURL, livechatEnvironment),
		tokenFn:    tokenFn,
	}

	return &Service{
		billingAPI:  a,
		storage:     storage,
		plans:       plans,
		returnURL:   returnUrl,
		masterOrgID: masterOrgID,
	}
}

func (s *Service) CreateRecurrentCharge(ctx context.Context, name string, price int, lcOrganizationID string) (string, error) {
	lcCharge, err := s.billingAPI.CreateRecurrentCharge(ctx, createRecurrentChargeParams{
		Name:      name,
		ReturnURL: s.returnURL,
		Price:     price,
		Test:      s.masterOrgID == lcOrganizationID,
		TrialDays: 0,
		Months:    1,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create recurrent charge via lc: %w", err)
	}

	if lcCharge == nil {
		return "", fmt.Errorf("failed to create recurrent charge via lc: charge is nil")
	}

	rawCharge, _ := json.Marshal(lcCharge)

	if err = s.storage.CreateCharge(ctx, Charge{
		LCOrganizationID: lcOrganizationID,
		ID:               lcCharge.ID,
		Type:             ChargeTypeRecurring,
		Payload:          rawCharge,
	}); err != nil {
		return "", fmt.Errorf("failed to create charge in database: %w", err)
	}

	return lcCharge.ID, nil
}

func (s *Service) SyncRecurrentCharge(ctx context.Context, id string) error {
	charge, err := s.storage.GetCharge(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get charge: %w", err)
	}

	if charge == nil {
		return fmt.Errorf("charge not found")
	}

	lcCharge, err := s.billingAPI.GetRecurrentCharge(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get recurrent charge: %w", err)
	}

	if err = s.storage.UpdateChargePayload(ctx, id, lcCharge.BaseCharge); err != nil {
		return fmt.Errorf("failed to update charge payload: %w", err)
	}

	return nil
}

func (s *Service) GetCharge(ctx context.Context, id string) (*Charge, error) {
	return s.storage.GetCharge(ctx, id)
}

func (s *Service) IsPremium(ctx context.Context, id string) (bool, error) {
	sub, err := s.storage.GetSubscriptionByOrganizationID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to get charge by installation id: %w", err)
	}

	return sub != nil && sub.IsActive(), nil
}

func (s *Service) CreateSubscription(ctx context.Context, lcOrganizationID string, planName string) error {
	plan := s.plans.GetPlan(planName)
	if plan == nil {
		return fmt.Errorf("plan not found")
	}

	charge, err := s.storage.GetChargeByOrganizationID(ctx, lcOrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get charge by organization id: %w", err)
	}

	if charge == nil {
		return fmt.Errorf("charge not found")
	}

	if err = s.storage.CreateSubscription(ctx, Subscription{
		ID:               xid.New().String(),
		Charge:           charge,
		LCOrganizationID: lcOrganizationID,
		PlanName:         planName,
	}); err != nil {
		return fmt.Errorf("failed to create subscription in database: %w", err)
	}

	return nil
}
