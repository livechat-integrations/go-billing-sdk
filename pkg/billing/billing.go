package billing

import (
	"context"
	"encoding/json"
	"fmt"
)

type billingAPI interface {
	CreateRecurrentCharge(ctx context.Context, params CreateRecurrentChargeParams) (*RecurrentCharge, error)
	GetRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error)
}

type Service struct {
	billingAPI    billingAPI
	storage       storage
	returnURL     string
	masterOrgID   string
	baseChatCount int64
}

func NewService(api billingAPI, storage storage, returnUrl, masterOrgID string, baseChatCount int64) *Service {
	return &Service{
		billingAPI:    api,
		storage:       storage,
		returnURL:     returnUrl,
		masterOrgID:   masterOrgID,
		baseChatCount: baseChatCount,
	}
}

func (s *Service) CreateRecurrentCharge(ctx context.Context, name string, price int, lcOrganizationID string) (string, error) {
	lcCharge, err := s.billingAPI.CreateRecurrentCharge(ctx, CreateRecurrentChargeParams{
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

	rawCharge, err := json.Marshal(lcCharge)
	if err != nil {
		return "", fmt.Errorf("failed to marshal charge: %w", err)
	}
	charge := InstallationCharge{
		InstallationID: lcOrganizationID,
		Charge: &Charge{
			ID:      lcCharge.ID,
			Type:    ChargeTypeRecurring,
			Payload: rawCharge,
		},
	}

	if err = s.storage.CreateCharge(ctx, charge); err != nil {
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

func (s *Service) GetCharge(ctx context.Context, id string) (*InstallationCharge, error) {
	return s.storage.GetCharge(ctx, id)
}

func (s *Service) IsValidSubscription(ctx context.Context, id string, chatsCount int64) (bool, error) {
	charge, err := s.storage.GetChargeByInstallationID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to get charge by installation id: %w", err)
	}

	if charge == nil || !charge.IsActive() {
		return chatsCount < s.baseChatCount, nil
	}

	return true, nil
}

func (s *Service) IsPremium(ctx context.Context, id string) (bool, error) {
	charge, err := s.storage.GetChargeByInstallationID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to get charge by installation id: %w", err)
	}

	return charge != nil && charge.IsActive(), nil
}
