package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

type LedgerInterface interface {
	CreateCharge(ctx context.Context, params CreateChargeParams) (string, error)
	CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (string, error)
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	GetTopUps(ctx context.Context, organizationID string) ([]TopUp, error)
	CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error
	ForceCancelTopUp(ctx context.Context, organizationID string, ID string) error
	CancelCharge(ctx context.Context, organizationID string, ID string) error
}

var (
	ErrChargeNotFound = fmt.Errorf("charge not found")
	ErrTopUpNotFound  = fmt.Errorf("top up not found")
)

type Service struct {
	xIdProvider common.XIdProviderInterface
	billingAPI  livechat.ApiInterface
	storage     Storage
	returnURL   string
	masterOrgID string
}

func NewService(xIdProvider common.XIdProviderInterface, httpClient *http.Client, livechatEnvironment string, tokenFn livechat.TokenFn, storage Storage, returnUrl, masterOrgID string) *Service {
	a := &livechat.Api{
		HttpClient: httpClient,
		ApiBaseURL: common.EnvURL(livechat.BillingAPIBaseURL, livechatEnvironment),
		TokenFn:    tokenFn,
	}

	return &Service{
		xIdProvider: xIdProvider,
		billingAPI:  a,
		storage:     storage,
		returnURL:   returnUrl,
		masterOrgID: masterOrgID,
	}
}

type CreateChargeParams struct {
	Test           bool    `json:"test"`
	Name           string  `json:"name"`
	Amount         float32 `json:"amount"`
	OrganizationID string  `json:"organizationId"`
}

func (s *Service) CreateCharge(ctx context.Context, params CreateChargeParams) (string, error) {
	event := toEvent(s.getUniqueID(), params.OrganizationID, EventActionCreateCharge, EventTypeInfo, params)
	charge := Charge{
		ID:               s.getUniqueID(),
		Amount:           params.Amount,
		Status:           ChargeStatusActive,
		LCOrganizationID: params.OrganizationID,
	}
	if err := s.storage.CreateCharge(ctx, charge); err != nil {
		event.Type = EventTypeError
		return "", s.toError(ctx, toErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create charge in database: %w", err),
		})
	}

	event.SetPayload(charge)
	_ = s.createEvent(ctx, event)

	return charge.ID, nil
}

type CreateTopUpRequestParams struct {
	Test           bool
	Name           string
	Amount         float32
	OrganizationID string
	Type           TopUpType
	Config         TopUpConfig
}

type TopUpConfig struct {
	ConfirmationUrl string  `json:"confirmationUrl"`
	TrialDays       *int    `json:"trialDays"`
	Months          *int    `json:"months"`
	ReturnUrl       *string `json:"returnUrl"`
}

func (s *Service) CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (string, error) {
	event := toEvent(s.getUniqueID(), params.OrganizationID, EventActionCreateTopUp, EventTypeInfo, params)
	isTest := params.Test || params.OrganizationID == s.masterOrgID
	config := ChargeConfig{
		ReturnUrl: &s.returnURL,
	}

	if params.Config.ReturnUrl != nil {
		config.ReturnUrl = params.Config.ReturnUrl
	}
	if params.Config.TrialDays != nil {
		config.TrialDays = params.Config.TrialDays
	}
	if params.Config.Months != nil {
		config.Months = params.Config.Months
	}

	chargeID, rawCharge, err := s.createBillingCharge(ctx, createBillingChargeParams{
		Test:           isTest,
		OrganizationID: params.OrganizationID,
		Name:           params.Name,
		Amount:         params.Amount,
		Type:           params.Type,
		Config:         config,
	})
	if err != nil {
		event.Type = EventTypeError
		return "", s.toError(ctx, toErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create top up billing charge: %w", err),
		})
	}
	if rawCharge == nil || chargeID == nil {
		event.Type = EventTypeError
		return "", s.toError(ctx, toErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create billing charge: empty charge id"),
		})
	}

	// TODO handle following in DPS webhook?
	// CurrentToppedUpAt: time.Time{},
	// NextTopUpAt:       time.Time{},

	topUp := TopUp{
		ID:               *chargeID,
		LCOrganizationID: params.OrganizationID,
		Status:           TopUpStatusPending,
		Amount:           params.Amount,
		Type:             params.Type,
		ConfirmationUrl:  params.Config.ConfirmationUrl,
		LCCharge:         *rawCharge,
	}
	err = s.storage.CreateTopUp(ctx, topUp)
	if err != nil {
		event.Type = EventTypeError
		return "", s.toError(ctx, toErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create database top up: %w", err),
		})
	}
	event.SetPayload(topUp)
	_ = s.createEvent(ctx, event)
	return *chargeID, nil
}

func (s *Service) GetBalance(ctx context.Context, organizationID string) (float32, error) {
	balance, err := s.storage.GetBalance(ctx, organizationID)
	if err != nil {
		return float32(0), fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (s *Service) GetTopUps(ctx context.Context, organizationID string) ([]TopUp, error) {
	return s.storage.GetTopUpsByOrganizationID(ctx, organizationID)
}

func (s *Service) CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error {
	event := toEvent(s.getUniqueID(), organizationID, EventActionCancelTopUp, EventTypeInfo, map[string]interface{}{"id": ID})
	topUp, err := s.storage.GetTopUpByIdAndType(ctx, ID, TopUpTypeRecurrent)
	if err != nil {
		event.Type = EventTypeError
		return s.toError(ctx, toErrorParams{
			event: event,
			err:   err,
		})
	}
	if topUp == nil {
		event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
		_ = s.createEvent(ctx, event)
		return ErrTopUpNotFound
	}

	_, err = s.billingAPI.CancelRecurrentCharge(ctx, ID)
	if err != nil {
		event.Type = EventTypeError
		return s.toError(ctx, toErrorParams{
			event: event,
			err:   err,
		})
	}

	err = s.storage.UpdateTopUpStatus(ctx, ID, TopUpStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
			_ = s.createEvent(ctx, event)
			return ErrTopUpNotFound
		}

		event.Type = EventTypeError
		return s.toError(ctx, toErrorParams{
			event: event,
			err:   err,
		})
	}

	event.SetPayload(map[string]interface{}{"id": ID, "result": "success"})
	_ = s.createEvent(ctx, event)
	return nil
}

func (s *Service) ForceCancelTopUp(ctx context.Context, organizationID string, ID string) error {
	event := toEvent(s.getUniqueID(), organizationID, EventActionForceCancelTopUp, EventTypeInfo, map[string]interface{}{"id": ID, "status": TopUpStatusCancelled})
	err := s.storage.UpdateTopUpStatus(ctx, ID, TopUpStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
			_ = s.createEvent(ctx, event)
			return ErrTopUpNotFound
		}

		event.Type = EventTypeError
		return s.toError(ctx, toErrorParams{
			event: event,
			err:   err,
		})
	}
	_ = s.createEvent(ctx, event)
	return nil
}

func (s *Service) CancelCharge(ctx context.Context, organizationID string, ID string) error {
	event := toEvent(s.getUniqueID(), organizationID, EventActionCancelCharge, EventTypeInfo, map[string]interface{}{"id": ID})
	err := s.storage.UpdateChargeStatus(ctx, ID, ChargeStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "charge not found"})
			_ = s.createEvent(ctx, event)
			return ErrChargeNotFound
		}
		event.Type = EventTypeError
		return s.toError(ctx, toErrorParams{
			event: event,
			err:   err,
		})
	}

	event.SetPayload(map[string]interface{}{"id": ID, "result": "success"})
	_ = s.createEvent(ctx, event)
	return nil
}

type createBillingChargeParams struct {
	Test           bool
	Name           string
	Amount         float32
	OrganizationID string
	Type           TopUpType
	Config         ChargeConfig
}

type ChargeConfig struct {
	TrialDays *int    `json:"trialDays"`
	Months    *int    `json:"months"`
	ReturnUrl *string `json:"returnUrl"`
}

func (s *Service) createBillingCharge(ctx context.Context, params createBillingChargeParams) (*string, *json.RawMessage, error) {
	isTest := params.Test || params.OrganizationID == s.masterOrgID
	returnUrl := s.returnURL
	if params.Config.ReturnUrl != nil {
		returnUrl = *params.Config.ReturnUrl
	}

	var rawCharge json.RawMessage
	var chargeID string
	switch params.Type {
	case TopUpTypeDirect:
		lcCharge, err := s.billingAPI.CreateDirectCharge(ctx, livechat.CreateDirectChargeParams{
			Name:      params.Name,
			ReturnURL: returnUrl,
			Price:     params.Amount,
			Test:      isTest,
		})

		if err != nil {
			return nil, nil, fmt.Errorf("failed to create direct charge via lc: %w", err)
		}

		if lcCharge == nil {
			return nil, nil, fmt.Errorf("failed to create direct charge via lc: charge is nil")
		}

		rawCharge, err = json.Marshal(lcCharge)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal lc direct charge: %w", err)
		}
		chargeID = lcCharge.ID
	case TopUpTypeRecurrent:
		if params.Config.Months == nil {
			return nil, nil, fmt.Errorf("failed to create recurrent charge v2 via lc: charge config months is nil")
		}
		recurrentChargeParams := livechat.CreateRecurrentChargeV2Params{
			Name:      params.Name,
			ReturnURL: returnUrl,
			Price:     params.Amount,
			Test:      isTest,
			Months:    *params.Config.Months,
		}
		if params.Config.TrialDays != nil {
			recurrentChargeParams.TrialDays = *params.Config.TrialDays
		}
		lcCharge, err := s.billingAPI.CreateRecurrentChargeV2(ctx, recurrentChargeParams)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to create recurrent charge v2 via lc: %w", err)
		}

		if lcCharge == nil {
			return nil, nil, fmt.Errorf("failed to create recurrent charge v2 via lc: charge is nil")
		}

		rawCharge, err = json.Marshal(lcCharge)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal lc recurrent charge: %w", err)
		}
		chargeID = lcCharge.ID
	}

	return &chargeID, &rawCharge, nil
}

func (s *Service) createEvent(ctx context.Context, event Event) error {
	err := s.storage.CreateEvent(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

type toErrorParams struct {
	event Event
	err   error
}

func (s *Service) toError(ctx context.Context, params toErrorParams) error {
	_ = s.createEvent(ctx, params.event)
	return fmt.Errorf("%s: %w", params.event.ID, params.err)
}

func toEvent(id string, organizationID string, action EventAction, eventType EventType, payload any) Event {
	event := Event{
		ID:               id,
		LCOrganizationID: organizationID,
		Type:             eventType,
		Action:           action,
		CreatedAt:        time.Time{},
	}
	jp, err := json.Marshal(payload)
	if err == nil {
		event.Payload = jp
	}

	return event
}

func (s *Service) getUniqueID() string {
	return s.xIdProvider.GenerateXId()
}
