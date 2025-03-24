package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/livechat-integrations/go-billing-sdk/pkg/common"
	"github.com/livechat-integrations/go-billing-sdk/pkg/common/livechat"
)

type LedgerInterface interface {
	CreateCharge(ctx context.Context, params CreateChargeParams) (string, error)
	CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (string, error)
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	GetTopUps(ctx context.Context, organizationID string) ([]TopUp, error)
	GetTopUpByID(ctx context.Context, ID string) (*TopUp, error)
	CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error
	ForceCancelTopUp(ctx context.Context, organizationID string, ID string) error
	CancelCharge(ctx context.Context, organizationID string, ID string) error
	GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error)
	ToError(ctx context.Context, params ToErrorParams) error
	ToEvent(ctx context.Context, organizationID string, action EventAction, eventType EventType, payload any) Event
	GetUniqueID() string
	CreateEvent(ctx context.Context, event Event) error
	UpdateTopUpStatus(ctx context.Context, organizationID string, ID string, status TopUpStatus) error
	SyncTopUp(ctx context.Context, organizationID string, ID string) (*TopUp, error)
}

var (
	ErrChargeNotFound = fmt.Errorf("charge not found")
	ErrTopUpNotFound  = fmt.Errorf("top up not found")
)

const (
	queryChargeIDKey = "ext_charge_id"
)

type (
	LedgerEventIDCtxKey        struct{}
	LedgerOrganizationIDCtxKey struct{}
)

type Service struct {
	idProvider  common.IdProviderInterface
	billingAPI  livechat.ApiInterface
	storage     Storage
	returnURL   string
	masterOrgID string
}

func NewService(idProvider common.IdProviderInterface, httpClient *http.Client, livechatEnvironment string, tokenFn livechat.TokenFn, storage Storage, returnUrl, masterOrgID string) *Service {
	a := &livechat.Api{
		HttpClient: httpClient,
		ApiBaseURL: common.EnvURL(livechat.BillingAPIBaseURL, livechatEnvironment),
		TokenFn:    tokenFn,
	}

	return &Service{
		idProvider:  idProvider,
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
	event := s.ToEvent(ctx, params.OrganizationID, EventActionCreateCharge, EventTypeInfo, params)
	charge := Charge{
		ID:               s.GetUniqueID(),
		Amount:           params.Amount,
		Status:           ChargeStatusActive,
		LCOrganizationID: params.OrganizationID,
	}
	if err := s.storage.CreateCharge(ctx, charge); err != nil {
		event.Type = EventTypeError
		return "", s.ToError(ctx, ToErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create charge in database: %w", err),
		})
	}

	event.SetPayload(charge)
	_ = s.CreateEvent(ctx, event)

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
	TrialDays *int    `json:"trialDays"`
	Months    *int    `json:"months"`
	ReturnUrl *string `json:"returnUrl"`
}

func (s *Service) CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (string, error) {
	event := s.ToEvent(ctx, params.OrganizationID, EventActionCreateTopUp, EventTypeInfo, params)
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

	cr, err := s.createBillingCharge(ctx, createBillingChargeParams{
		Test:           isTest,
		OrganizationID: params.OrganizationID,
		Name:           params.Name,
		Amount:         params.Amount,
		Type:           params.Type,
		Config:         config,
	})
	if err != nil {
		event.Type = EventTypeError
		return "", s.ToError(ctx, ToErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create top up billing charge: %w", err),
		})
	}
	if cr.RawCharge == nil || cr.ChargeID == nil || cr.ConfirmationUrl == nil {
		event.Type = EventTypeError
		return "", s.ToError(ctx, ToErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create billing charge: empty charge id"),
		})
	}

	topUp := TopUp{
		ID:               *cr.ChargeID,
		LCOrganizationID: params.OrganizationID,
		Status:           TopUpStatusPending,
		Amount:           params.Amount,
		Type:             params.Type,
		ConfirmationUrl:  *cr.ConfirmationUrl,
		LCCharge:         *cr.RawCharge,
	}
	err = s.storage.CreateTopUp(ctx, topUp)
	if err != nil {
		event.Type = EventTypeError
		return "", s.ToError(ctx, ToErrorParams{
			event: event,
			err:   fmt.Errorf("failed to create database top up: %w", err),
		})
	}
	event.SetPayload(topUp)
	_ = s.CreateEvent(ctx, event)
	return topUp.ID, nil
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

func (s *Service) GetTopUpByID(ctx context.Context, ID string) (*TopUp, error) {
	return s.storage.GetTopUpByID(ctx, ID)
}

func (s *Service) CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error {
	event := s.ToEvent(ctx, organizationID, EventActionCancelTopUp, EventTypeInfo, map[string]interface{}{"id": ID})
	topUp, err := s.storage.GetTopUpByIDAndType(ctx, ID, TopUpTypeRecurrent)
	if err != nil {
		event.Type = EventTypeError
		return s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}
	if topUp == nil {
		event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
		_ = s.CreateEvent(ctx, event)
		return ErrTopUpNotFound
	}

	_, err = s.billingAPI.CancelRecurrentCharge(ctx, ID)
	if err != nil {
		event.Type = EventTypeError
		return s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}

	err = s.storage.UpdateTopUpStatus(ctx, ID, TopUpStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
			_ = s.CreateEvent(ctx, event)
			return ErrTopUpNotFound
		}

		event.Type = EventTypeError
		return s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}

	event.SetPayload(map[string]interface{}{"id": ID, "result": "success"})
	_ = s.CreateEvent(ctx, event)
	return nil
}

func (s *Service) ForceCancelTopUp(ctx context.Context, organizationID string, ID string) error {
	return s.UpdateTopUpStatus(ctx, organizationID, ID, TopUpStatusCancelled)
}

func (s *Service) UpdateTopUpStatus(ctx context.Context, organizationID string, ID string, status TopUpStatus) error {
	event := s.ToEvent(ctx, organizationID, EventActionUpdateTopUpStatus, EventTypeInfo, map[string]interface{}{"id": ID, "status": status})
	err := s.storage.UpdateTopUpStatus(ctx, ID, status)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
			_ = s.CreateEvent(ctx, event)
			return ErrTopUpNotFound
		}

		event.Type = EventTypeError
		return s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}
	_ = s.CreateEvent(ctx, event)
	return nil
}

func (s *Service) CancelCharge(ctx context.Context, organizationID string, ID string) error {
	event := s.ToEvent(ctx, organizationID, EventActionCancelCharge, EventTypeInfo, map[string]interface{}{"id": ID})
	err := s.storage.UpdateChargeStatus(ctx, ID, ChargeStatusCancelled)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "charge not found"})
			_ = s.CreateEvent(ctx, event)
			return ErrChargeNotFound
		}
		event.Type = EventTypeError
		return s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}

	event.SetPayload(map[string]interface{}{"id": ID, "result": "success"})
	_ = s.CreateEvent(ctx, event)
	return nil
}

func (s *Service) GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error) {
	return s.storage.GetTopUpsByOrganizationIDAndStatus(ctx, organizationID, status)
}

func (s *Service) SyncTopUp(ctx context.Context, organizationID string, ID string) (*TopUp, error) {
	event := s.ToEvent(ctx, organizationID, EventActionSyncTopUp, EventTypeInfo, map[string]interface{}{"id": ID})
	var baseCharge livechat.BaseChargeV2
	var fullCharge any
	var chargeType TopUpType
	status := TopUpStatusPending
	topUp := TopUp{
		LCOrganizationID: organizationID,
	}
	isDirect := false
	isRecurrent := false

	eg := errgroup.Group{}
	eg.Go(func() error {
		c, err := s.billingAPI.GetDirectCharge(ctx, ID)
		if err != nil {
			return err
		}
		if c == nil {
			return nil
		}
		fullCharge = c
		baseCharge = c.BaseChargeV2
		chargeType = TopUpTypeDirect
		switch baseCharge.Status {
		case "success":
			status = TopUpStatusActive
		case "failed":
			status = TopUpStatusCancelled
		case "declined":
			status = TopUpStatusCancelled
		}
		isDirect = true
		return nil
	})
	eg.Go(func() error {
		c, err := s.billingAPI.GetRecurrentChargeV2(ctx, ID)
		if err != nil {
			return err
		}
		if c == nil {
			return nil
		}
		fullCharge = c
		baseCharge = c.BaseChargeV2
		chargeType = TopUpTypeRecurrent
		switch baseCharge.Status {
		case "active":
			status = TopUpStatusActive
		case "cancelled":
			status = TopUpStatusCancelled
		case "declined":
			status = TopUpStatusCancelled
		}
		topUp.CurrentToppedUpAt = c.CurrentChargeAt
		topUp.NextTopUpAt = c.NextChargeAt
		isRecurrent = true
		return nil
	})
	if err := eg.Wait(); err != nil && !isDirect && !isRecurrent {
		event.Type = EventTypeError
		return nil, s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}
	if isDirect && isRecurrent {
		event.Type = EventTypeError
		return nil, s.ToError(ctx, ToErrorParams{
			event: event,
			err:   fmt.Errorf("charge conflict"),
		})
	}

	if fullCharge == nil {
		event.Type = EventTypeError
		return nil, s.ToError(ctx, ToErrorParams{
			event: event,
			err:   fmt.Errorf("charge not found"),
		})
	}

	u, err := url.Parse(baseCharge.ReturnURL)
	if err != nil {
		event.Type = EventTypeError
		return nil, s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}
	id := u.Query().Get(queryChargeIDKey)
	if id == "" {
		id = baseCharge.ID
	}

	p, err := json.Marshal(fullCharge)
	if err != nil {
		event.Type = EventTypeError
		return nil, s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}
	topUp.ID = id
	topUp.Type = chargeType
	topUp.Status = status
	topUp.LCCharge = p
	topUp.Amount = baseCharge.Price
	topUp.ConfirmationUrl = baseCharge.ConfirmationURL

	uTopUp, err := s.storage.UpsertTopUp(ctx, topUp)
	if err != nil {
		event.Type = EventTypeError
		return nil, s.ToError(ctx, ToErrorParams{
			event: event,
			err:   err,
		})
	}
	event.SetPayload(topUp)
	_ = s.CreateEvent(ctx, event)

	return uTopUp, nil
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

type createBillingChargeResult struct {
	RawCharge       *[]byte `json:"charge"`
	ConfirmationUrl *string `json:"confirmationUrl"`
	ChargeID        *string `json:"chargeId"`
}

func (s *Service) createBillingCharge(ctx context.Context, params createBillingChargeParams) (*createBillingChargeResult, error) {
	isTest := params.Test || params.OrganizationID == s.masterOrgID
	returnUrl := s.returnURL
	if params.Config.ReturnUrl != nil {
		returnUrl = *params.Config.ReturnUrl
	}

	var result createBillingChargeResult
	switch params.Type {
	case TopUpTypeDirect:
		lcCharge, err := s.billingAPI.CreateDirectCharge(ctx, livechat.CreateDirectChargeParams{
			Name:      params.Name,
			ReturnURL: returnUrl,
			Price:     params.Amount,
			Test:      isTest,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to create direct charge via lc: %w", err)
		}

		if lcCharge == nil {
			return nil, fmt.Errorf("failed to create direct charge via lc: charge is nil")
		}

		rawCharge, err := json.Marshal(lcCharge)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal lc direct charge: %w", err)
		}
		result.RawCharge = &rawCharge
		result.ConfirmationUrl = &lcCharge.ConfirmationURL
		result.ChargeID = &lcCharge.ID
	case TopUpTypeRecurrent:
		if params.Config.Months == nil {
			return nil, fmt.Errorf("failed to create recurrent charge v2 via lc: charge config months is nil")
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
			return nil, fmt.Errorf("failed to create recurrent charge v2 via lc: %w", err)
		}

		if lcCharge == nil {
			return nil, fmt.Errorf("failed to create recurrent charge v2 via lc: charge is nil")
		}

		rawCharge, err := json.Marshal(lcCharge)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal lc recurrent charge: %w", err)
		}
		result.RawCharge = &rawCharge
		result.ConfirmationUrl = &lcCharge.ConfirmationURL
		result.ChargeID = &lcCharge.ID
	}

	return &result, nil
}

func (s *Service) CreateEvent(ctx context.Context, event Event) error {
	err := s.storage.CreateEvent(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

type ToErrorParams struct {
	event Event
	err   error
}

func (s *Service) ToError(ctx context.Context, params ToErrorParams) error {
	params.event.Error = params.err.Error()
	_ = s.CreateEvent(ctx, params.event)
	return fmt.Errorf("%s: %w", params.event.ID, params.err)
}

func (s *Service) ToEvent(ctx context.Context, organizationID string, action EventAction, eventType EventType, payload any) Event {
	id, ok := ctx.Value(LedgerEventIDCtxKey{}).(string)
	if !ok {
		id = s.idProvider.GenerateId()
	}

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

func (s *Service) GetUniqueID() string {
	return s.idProvider.GenerateId()
}
