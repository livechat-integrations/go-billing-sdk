package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/v2/common"
	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/livechat-integrations/go-billing-sdk/v2/pkg/events"
)

type LedgerInterface interface {
	GetOperations(ctx context.Context, organizationID string) ([]Operation, error)
	CreateCharge(ctx context.Context, params CreateChargeParams) (string, error)
	TopUp(ctx context.Context, topUp TopUp) (string, error)
	CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (*TopUp, error)
	GetBalance(ctx context.Context, organizationID string) (float32, error)
	GetTopUps(ctx context.Context, organizationID string) ([]TopUp, error)
	CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error
	ForceCancelTopUp(ctx context.Context, topUp TopUp) error
	GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error)
	GetTopUpByIDAndOrganizationID(ctx context.Context, organizationID string, ID string) (*TopUp, error)
	SyncTopUp(ctx context.Context, topUp TopUp) (*TopUp, error)
	SyncOrCancelTopUpRequests(ctx context.Context) error
	AddFunds(ctx context.Context, Amount float32, OrganizationID, Namespace string) error
	RecentlyAddedFunds(ctx context.Context, OrganizationID, Namespace string) (*Operation, error)
}

var (
	ErrChargeNotFound = fmt.Errorf("charge not found")
	ErrTopUpNotFound  = fmt.Errorf("top up not found")
)

type (
	LedgerEventIDCtxKey        struct{}
	LedgerOrganizationIDCtxKey struct{}
)

type Service struct {
	idProvider   events.IdProviderInterface
	billingAPI   livechat.ApiInterface
	eventService events.EventService
	storage      Storage
	returnURL    string
	masterOrgID  string
}

func NewService(eventService events.EventService, idProvider events.IdProviderInterface, httpClient *http.Client, livechatEnvironment string, tokenFn common.TokenFn, storage Storage, returnUrl, masterOrgID string) *Service {
	a := &livechat.Api{
		HttpClient: httpClient,
		ApiBaseURL: events.EnvURL(livechat.BillingAPIBaseURL, livechatEnvironment),
		TokenFn:    tokenFn,
	}

	return &Service{
		idProvider:   idProvider,
		billingAPI:   a,
		eventService: eventService,
		storage:      storage,
		returnURL:    returnUrl,
		masterOrgID:  masterOrgID,
	}
}

type CreateChargeParams struct {
	Test           bool    `json:"test"`
	Name           string  `json:"name"`
	Amount         float32 `json:"amount"`
	OrganizationID string  `json:"organizationId"`
}

func (s *Service) GetOperations(ctx context.Context, organizationID string) ([]Operation, error) {
	return s.storage.GetLedgerOperations(ctx, organizationID)
}

func (s *Service) CreateCharge(ctx context.Context, params CreateChargeParams) (string, error) {
	operation := Operation{
		ID:               s.idProvider.GenerateId(),
		Amount:           -params.Amount,
		LCOrganizationID: params.OrganizationID,
	}
	_, err := s.createOperation(ctx, operation)
	if err != nil {
		return "", err
	}
	return operation.ID, nil
}

func (s *Service) TopUp(ctx context.Context, topUp TopUp) (string, error) {
	event := s.eventService.ToEvent(ctx, topUp.LCOrganizationID, events.EventActionTopUp, events.EventTypeInfo, topUp)
	dbTopUp, err := s.storage.GetTopUpByIDAndType(ctx, GetTopUpByIDAndTypeParams{
		ID:   topUp.ID,
		Type: topUp.Type,
	})
	if err != nil {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}
	if dbTopUp == nil {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("no existing top up in database"),
		})
	}
	if !(dbTopUp.Status == TopUpStatusSuccess || dbTopUp.Status == TopUpStatusActive) {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("top up has wrong status: %s", dbTopUp.Status),
		})
	}

	if dbTopUp.Type == TopUpTypeRecurrent {
		var charge livechat.RecurrentCharge
		err = json.Unmarshal(dbTopUp.LCCharge, &charge)
		if err != nil {
			event.Type = events.EventTypeError
			return "", s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}
		if charge.NextChargeAt == nil || charge.CurrentChargeAt == nil {
			event.Type = events.EventTypeError
			return "", s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("no charge at current time"),
			})
		}
		dbTopUp.NextTopUpAt = charge.NextChargeAt
		dbTopUp.CurrentToppedUpAt = charge.CurrentChargeAt
		dbTopUp, err = s.storage.UpsertTopUp(ctx, *dbTopUp)
		if err != nil {
			event.Type = events.EventTypeError
			return "", s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}
		if dbTopUp == nil {
			event.Type = events.EventTypeError
			return "", s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("upsert top up error"),
			})
		}
	}

	id := dbTopUp.ID
	if dbTopUp.Type == TopUpTypeRecurrent && dbTopUp.CurrentToppedUpAt != nil {
		id = fmt.Sprintf("%s-%d", id, dbTopUp.CurrentToppedUpAt.UnixMicro())
	}
	operation := Operation{
		ID:               id,
		Amount:           dbTopUp.Amount,
		LCOrganizationID: dbTopUp.LCOrganizationID,
		Payload:          dbTopUp.LCCharge,
	}
	_, err = s.createOperation(ctx, operation)
	if err != nil {
		event.Type = events.EventTypeError
		return "", s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}

	_ = s.eventService.CreateEvent(ctx, event)

	return operation.ID, nil
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

func (s *Service) CreateTopUpRequest(ctx context.Context, params CreateTopUpRequestParams) (*TopUp, error) {
	event := s.eventService.ToEvent(ctx, params.OrganizationID, events.EventActionCreateTopUp, events.EventTypeInfo, params)
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
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create top up billing charge: %w", err),
		})
	}
	if cr.RawCharge == nil || cr.ChargeID == nil || cr.ConfirmationUrl == nil {
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create billing charge: empty charge id"),
		})
	}

	topUp := TopUp{
		ID:               *cr.ChargeID,
		LCOrganizationID: params.OrganizationID,
		Status:           TopUpStatusPending,
		Amount:           float32(params.Amount),
		Type:             params.Type,
		ConfirmationUrl:  *cr.ConfirmationUrl,
		LCCharge:         *cr.RawCharge,
	}
	if cr.NextChargeAt != nil {
		topUp.NextTopUpAt = cr.NextChargeAt
	}
	if cr.CurrentChargeAt != nil {
		topUp.CurrentToppedUpAt = cr.CurrentChargeAt
	}

	tu, err := s.storage.UpsertTopUp(ctx, topUp)
	if err != nil {
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create database top up: %w", err),
		})
	}
	event.SetPayload(tu)
	_ = s.eventService.CreateEvent(ctx, event)
	return tu, nil
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

func (s *Service) AddFunds(ctx context.Context, Amount float32, OrganizationID, Namespace string) error {
	event := s.eventService.ToEvent(ctx, OrganizationID, events.EventActionAddFunds, events.EventTypeInfo, map[string]interface{}{"amount": Amount, "namespace": Namespace})
	key := getFundsKey(Namespace, OrganizationID)
	operation, err := s.storage.GetLedgerOperation(ctx, GetLedgerOperationParams{
		ID:             key,
		OrganizationID: OrganizationID,
	})
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed get ledger operation from database: %w", err),
		})
	}
	if operation != nil {
		event.SetPayload(map[string]interface{}{"id": key, "amount": Amount, "namespace": Namespace, "result": "already exists"})
		_ = s.eventService.CreateEvent(ctx, event)
		return nil
	}
	operation = &Operation{
		ID:               key,
		LCOrganizationID: OrganizationID,
		Amount:           Amount,
	}
	_, err = s.createOperation(ctx, *operation)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}
	_ = s.eventService.CreateEvent(ctx, event)

	return nil
}

func (s *Service) RecentlyAddedFunds(ctx context.Context, OrganizationID, Namespace string) (*Operation, error) {
	key := getFundsKey(Namespace, OrganizationID)
	operation, err := s.storage.GetLedgerOperation(ctx, GetLedgerOperationParams{
		ID:             key,
		OrganizationID: OrganizationID,
	})
	if err != nil {
		return nil, err
	}
	if operation != nil {
		return operation, nil
	}
	return nil, nil
}

func (s *Service) CancelTopUpRequest(ctx context.Context, organizationID string, ID string) error {
	event := s.eventService.ToEvent(ctx, organizationID, events.EventActionCancelRecurrentTopUp, events.EventTypeInfo, map[string]interface{}{"id": ID})
	topUp, err := s.storage.GetTopUpByIDAndType(ctx, GetTopUpByIDAndTypeParams{
		ID:   ID,
		Type: TopUpTypeRecurrent,
	})
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}
	if topUp == nil {
		event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
		_ = s.eventService.CreateEvent(ctx, event)
		return ErrTopUpNotFound
	}

	_, err = s.billingAPI.CancelRecurrentCharge(ctx, ID)
	if err != nil {
		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}

	err = s.storage.UpdateTopUpStatus(ctx, UpdateTopUpStatusParams{
		ID:     ID,
		Status: TopUpStatusCancelled,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": ID, "result": "top up not found"})
			_ = s.eventService.CreateEvent(ctx, event)
			return ErrTopUpNotFound
		}

		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}

	event.SetPayload(map[string]interface{}{"id": ID, "result": "success"})
	_ = s.eventService.CreateEvent(ctx, event)
	return nil
}

func (s *Service) ForceCancelTopUp(ctx context.Context, topUp TopUp) error {
	event := s.eventService.ToEvent(ctx, topUp.LCOrganizationID, events.EventActionForceCancelCharge, events.EventTypeInfo, map[string]interface{}{"id": topUp.ID, "status": TopUpStatusCancelled})
	err := s.storage.UpdateTopUpStatus(ctx, UpdateTopUpStatusParams{
		ID:     topUp.ID,
		Status: TopUpStatusCancelled,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			event.SetPayload(map[string]interface{}{"id": topUp.ID, "result": "top up not found"})
			_ = s.eventService.CreateEvent(ctx, event)
			return ErrTopUpNotFound
		}

		event.Type = events.EventTypeError
		return s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}
	_ = s.eventService.CreateEvent(ctx, event)
	return nil
}

func (s *Service) GetTopUpsByOrganizationIDAndStatus(ctx context.Context, organizationID string, status TopUpStatus) ([]TopUp, error) {
	return s.storage.GetTopUpsByOrganizationIDAndStatus(ctx, organizationID, status)
}

func (s *Service) GetTopUpByIDAndOrganizationID(ctx context.Context, organizationID string, ID string) (*TopUp, error) {
	return s.storage.GetTopUpByIDAndOrganizationID(ctx, organizationID, ID)
}

func (s *Service) SyncTopUp(ctx context.Context, topUp TopUp) (*TopUp, error) {
	event := s.eventService.ToEvent(ctx, topUp.LCOrganizationID, events.EventActionSyncTopUp, events.EventTypeInfo, topUp)
	var baseCharge livechat.BaseCharge
	var fullCharge any

	switch topUp.Type {
	case TopUpTypeDirect:
		c, err := s.billingAPI.GetDirectCharge(ctx, topUp.ID)
		if err != nil {
			event.Type = events.EventTypeError
			return nil, s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}
		if c == nil {
			event.Type = events.EventTypeError
			return nil, s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("failed to get LC API direct top up by id: %s", topUp.ID),
			})
		}
		fullCharge = c
		baseCharge = c.BaseCharge
		switch baseCharge.Status {
		case "success":
			topUp.Status = TopUpStatusSuccess
		case "accepted":
			topUp.Status = TopUpStatusAccepted
		case "processed":
			topUp.Status = TopUpStatusProcessing
		case "failed":
			topUp.Status = TopUpStatusFailed
		case "cancelled":
			topUp.Status = TopUpStatusCancelled
		case "declined":
			topUp.Status = TopUpStatusDeclined
		case "frozen":
			topUp.Status = TopUpStatusFrozen
		}
	case TopUpTypeRecurrent:
		c, err := s.billingAPI.GetRecurrentCharge(ctx, topUp.ID)
		if err != nil {
			event.Type = events.EventTypeError
			return nil, s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   err,
			})
		}
		if c == nil {
			event.Type = events.EventTypeError
			return nil, s.eventService.ToError(ctx, events.ToErrorParams{
				Event: event,
				Err:   fmt.Errorf("failed to get LC API recurrent top up by id: %s", topUp.ID),
			})
		}
		fullCharge = c
		baseCharge = c.BaseCharge
		switch baseCharge.Status {
		case "active":
			topUp.Status = TopUpStatusActive
		case "processed":
			topUp.Status = TopUpStatusProcessing
		case "accepted":
			topUp.Status = TopUpStatusAccepted
		case "cancelled":
			topUp.Status = TopUpStatusCancelled
		case "declined":
			topUp.Status = TopUpStatusDeclined
		case "frozen":
			topUp.Status = TopUpStatusFrozen
		case "past_due":
			topUp.Status = TopUpStatusPastDue
		}
	}

	if fullCharge == nil {
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("charge not found"),
		})
	}

	p, err := json.Marshal(fullCharge)
	if err != nil {
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}
	topUp.LCCharge = p
	if baseCharge.Price > 0 {
		topUp.Amount = float32(baseCharge.Price) / 100
	}
	topUp.ConfirmationUrl = baseCharge.ConfirmationURL

	uTopUp, err := s.storage.UpsertTopUp(ctx, topUp)
	if err != nil {
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   err,
		})
	}
	event.SetPayload(topUp)
	_ = s.eventService.CreateEvent(ctx, event)

	return uTopUp, nil
}

func (s *Service) SyncOrCancelTopUpRequests(ctx context.Context) error {
	topUps, err := s.storage.GetTopUpsByTypeWhereStatusNotIn(ctx, GetTopUpsByTypeWhereStatusNotInParams{
		Type: TopUpTypeDirect,
		Statuses: []TopUpStatus{
			TopUpStatusSuccess,
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined,
		},
	})
	if err != nil {
		return err
	}

	err = s.syncOrCancelDirectTopUpRequests(ctx, topUps)
	if err != nil {
		return err
	}

	topUps, err = s.storage.GetRecurrentTopUpsWhereStatusNotIn(ctx, []TopUpStatus{
		TopUpStatusCancelled,
		TopUpStatusFailed,
		TopUpStatusDeclined,
	})
	if err != nil {
		return err
	}

	err = s.syncOrCancelRecurrentTopUpRequests(ctx, topUps)
	if err != nil {
		return err
	}

	topUps, err = s.storage.GetDirectTopUpsWithoutOperations(ctx)
	if err != nil {
		return err
	}
	for _, topUp := range topUps {
		organizationCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, topUp.LCOrganizationID)
		organizationCtx = context.WithValue(organizationCtx, LedgerEventIDCtxKey{}, s.idProvider.GenerateId())
		_, err := s.TopUp(organizationCtx, topUp)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) syncOrCancelDirectTopUpRequests(ctx context.Context, topUps []TopUp) error {
	for _, topUp := range topUps {
		organizationCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, topUp.LCOrganizationID)
		organizationCtx = context.WithValue(organizationCtx, LedgerEventIDCtxKey{}, s.idProvider.GenerateId())
		tu, err := s.SyncTopUp(organizationCtx, topUp)
		if err != nil {
			return err
		}

		switch tu.Status {
		case TopUpStatusSuccess,
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined:
			// do nothing

		case TopUpStatusAccepted:
			_, err = s.billingAPI.ActivateDirectCharge(organizationCtx, tu.ID)
			if err != nil {
				return err
			}
			if err = s.eventService.CreateEvent(organizationCtx, s.eventService.ToEvent(organizationCtx, topUp.LCOrganizationID, events.EventActionActivateCharge, events.EventTypeInfo, tu)); err != nil {
				return err
			}
		default:
			monthAgo := time.Now().AddDate(0, -1, 0)
			if tu.Type == TopUpTypeDirect && monthAgo.After(tu.CreatedAt) {
				err = s.ForceCancelTopUp(organizationCtx, *tu)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Service) syncOrCancelRecurrentTopUpRequests(ctx context.Context, topUps []TopUp) error {
	for _, topUp := range topUps {
		organizationCtx := context.WithValue(ctx, LedgerOrganizationIDCtxKey{}, topUp.LCOrganizationID)
		organizationCtx = context.WithValue(organizationCtx, LedgerEventIDCtxKey{}, s.idProvider.GenerateId())
		tu, err := s.SyncTopUp(organizationCtx, topUp)
		if err != nil {
			return err
		}

		switch tu.Status {
		case TopUpStatusActive,
			TopUpStatusCancelled,
			TopUpStatusFailed,
			TopUpStatusDeclined:
			// do nothing

		case TopUpStatusAccepted,
			TopUpStatusFrozen:
			_, err = s.billingAPI.ActivateRecurrentCharge(organizationCtx, tu.ID)
			if err != nil {
				return err
			}
			if err = s.eventService.CreateEvent(organizationCtx, s.eventService.ToEvent(organizationCtx, topUp.LCOrganizationID, events.EventActionActivateCharge, events.EventTypeInfo, tu)); err != nil {
				return err
			}
		default:
			monthAgo := time.Now().AddDate(0, -1, 0)
			if tu.Type == TopUpTypeRecurrent && tu.CurrentToppedUpAt != nil && monthAgo.After(*tu.CurrentToppedUpAt) {
				err = s.ForceCancelTopUp(organizationCtx, *tu)
				if err != nil {
					return err
				}
			}
		}
	}

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

type createBillingChargeResult struct {
	RawCharge       *[]byte    `json:"charge"`
	ConfirmationUrl *string    `json:"confirmationUrl"`
	ChargeID        *string    `json:"chargeId"`
	NextChargeAt    *time.Time `json:"nextChargeAt"`
	CurrentChargeAt *time.Time `json:"currentChargeAt"`
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
		// multiply price by 100 because LC is using integer (1 = 1 cent)
		lcCharge, err := s.billingAPI.CreateDirectCharge(ctx, livechat.CreateDirectChargeParams{
			Name:      params.Name,
			ReturnURL: returnUrl,
			Price:     int(params.Amount * 100),
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
			return nil, fmt.Errorf("failed to create recurrent charge V3 via lc: charge config months is nil")
		}
		// multiply price by 100 because LC is using integer (1 = 1 cent)
		recurrentChargeParams := livechat.CreateRecurrentChargeParams{
			Name:      params.Name,
			ReturnURL: returnUrl,
			Price:     int(params.Amount * 100),
			Test:      isTest,
			Months:    *params.Config.Months,
		}
		if params.Config.TrialDays != nil {
			recurrentChargeParams.TrialDays = *params.Config.TrialDays
		}
		lcCharge, err := s.billingAPI.CreateRecurrentCharge(ctx, recurrentChargeParams)

		if err != nil {
			return nil, fmt.Errorf("failed to create recurrent charge V3 via lc: %w", err)
		}
		if lcCharge == nil {
			return nil, fmt.Errorf("failed to create recurrent charge V3 via lc: charge is nil")
		}

		rawCharge, err := json.Marshal(lcCharge)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal lc recurrent charge: %w", err)
		}
		result.RawCharge = &rawCharge
		result.ConfirmationUrl = &lcCharge.ConfirmationURL
		result.ChargeID = &lcCharge.ID
		result.NextChargeAt = lcCharge.NextChargeAt
		result.CurrentChargeAt = lcCharge.CurrentChargeAt
	}

	return &result, nil
}

func (s *Service) createOperation(ctx context.Context, operation Operation) (*Operation, error) {
	event := s.eventService.ToEvent(ctx, operation.LCOrganizationID, events.EventActionCreateOperation, events.EventTypeInfo, operation)
	err := s.storage.CreateLedgerOperation(ctx, operation)
	if err != nil {
		event.Type = events.EventTypeError
		return nil, s.eventService.ToError(ctx, events.ToErrorParams{
			Event: event,
			Err:   fmt.Errorf("failed to create ledger operation in database: %w", err),
		})
	}
	_ = s.eventService.CreateEvent(ctx, event)
	return &operation, nil
}

func getFundsKey(namespace, organizationID string) string {
	return fmt.Sprintf("add-funds-%s-%s", namespace, organizationID)
}
