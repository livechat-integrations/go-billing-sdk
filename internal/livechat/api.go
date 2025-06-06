package livechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/v2/common"
)

const BillingAPIBaseURL = "https://billing.livechatinc.com"

const (
	RecurrentChargeStatusPending   ChargeStatus = "pending"
	RecurrentChargeStatusAccepted  ChargeStatus = "accepted"
	RecurrentChargeStatusActive    ChargeStatus = "active"
	RecurrentChargeStatusCancelled ChargeStatus = "cancelled"
	RecurrentChargeStatusFrozen    ChargeStatus = "frozen"
	RecurrentChargeStatusPastDue   ChargeStatus = "past_due"

	DirectChargeStatusPending    ChargeStatus = "pending"
	DirectChargeStatusAccepted   ChargeStatus = "accepted"
	DirectChargeStatusDeclined   ChargeStatus = "declined"
	DirectChargeStatusProcessing ChargeStatus = "processed"
	DirectChargeStatusFailed     ChargeStatus = "failed"
	DirectChargeStatusSuccess    ChargeStatus = "success"
)

type ChargeStatus string
type BaseCharge struct {
	ID                  string       `json:"id"`
	BuyerLicenseID      int          `json:"buyer_license_id"`
	BuyerEntityID       string       `json:"buyer_entity_id"`
	BuyerOrganizationID string       `json:"buyer_organization_id"`
	BuyerAccountID      string       `json:"buyer_account_id"`
	SellerClientID      string       `json:"seller_client_id"`
	OrderClientID       string       `json:"order_client_id"`
	OrderLicenseID      string       `json:"order_license_id"`
	OrderEntityID       string       `json:"order_entity_id"`
	OrderOrganizationID string       `json:"order_organization_id"`
	Name                string       `json:"name"`
	Price               int          `json:"price"`
	ReturnURL           string       `json:"return_url"`
	Test                bool         `json:"test"`
	PerAccount          bool         `json:"per_account"`
	Status              ChargeStatus `json:"status"`
	ConfirmationURL     string       `json:"confirmation_url"`
	CommissionPercent   int          `json:"commission_percent"`
	UpdatedAt           *time.Time   `json:"updated_at"`
	CreatedAt           *time.Time   `json:"created_at"`
}

type DirectCharge struct {
	BaseCharge
	Quantity int `json:"quantity"`
}

type RecurrentCharge struct {
	BaseCharge
	TrialDays       int        `json:"trial_days"`
	Months          int        `json:"months"`
	TrialEndsAt     *time.Time `json:"trial_ends_at"`
	CancelledAt     *time.Time `json:"cancelled_at"`
	CurrentChargeAt *time.Time `json:"current_charge_at"`
	NextChargeAt    *time.Time `json:"next_charge_at"`
}

type ApiInterface interface {
	CreateDirectCharge(ctx context.Context, params CreateDirectChargeParams) (*DirectCharge, error)
	GetDirectCharge(ctx context.Context, id string) (*DirectCharge, error)
	CreateRecurrentCharge(ctx context.Context, params CreateRecurrentChargeParams) (*RecurrentCharge, error)
	GetRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error)
	CancelRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error)
	ActivateRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error)
	ActivateDirectCharge(ctx context.Context, id string) (*DirectCharge, error)
}

type httpCaller interface {
	Do(req *http.Request) (*http.Response, error)
}

type Api struct {
	HttpClient httpCaller
	ApiBaseURL string
	TokenFn    common.TokenFn
}

type CreateDirectChargeParams struct {
	Name              string
	ReturnURL         string
	Price             int
	Test              bool
	CommissionPercent *int
	Payload           []byte
}

func (a *Api) CreateDirectCharge(ctx context.Context, params CreateDirectChargeParams) (*DirectCharge, error) {
	type payload struct {
		Name              string `json:"name"`
		Price             int    `json:"price"`
		ReturnURL         string `json:"return_url"`
		Test              bool   `json:"test"`
		Quantity          int    `json:"quantity"`
		CommissionPercent *int   `json:"commission_percent,omitempty"`
		Payload           []byte `json:"payload"`
	}

	resp, err := a.call(ctx, "POST", "/v3/direct_charge/livechat", payload{
		Name:              params.Name,
		Price:             params.Price,
		ReturnURL:         params.ReturnURL,
		Test:              params.Test,
		Quantity:          1,
		CommissionPercent: params.CommissionPercent,
		Payload:           params.Payload,
	})
	if err != nil {
		return nil, err
	}

	return asDirectCharge(resp)
}

func (a *Api) GetDirectCharge(ctx context.Context, id string) (*DirectCharge, error) {
	resp, err := a.call(ctx, "GET", "/v3/direct_charge/livechat/"+id, nil)
	if err != nil {
		return nil, err
	}

	return asDirectCharge(resp)
}

type CreateRecurrentChargeParams struct {
	Name              string
	ReturnURL         string
	Price             int
	Test              bool
	TrialDays         int
	Months            int
	CommissionPercent *int
}

func (a *Api) CreateRecurrentCharge(ctx context.Context, params CreateRecurrentChargeParams) (*RecurrentCharge, error) {
	type payload struct {
		Name              string `json:"name"`
		Price             int    `json:"price"`
		ReturnURL         string `json:"return_url"`
		Test              bool   `json:"test"`
		TrialDays         int    `json:"trial_days"`
		Months            int    `json:"months,omitempty"`
		CommissionPercent *int   `json:"commission_percent,omitempty"`
	}
	resp, err := a.call(ctx, "POST", "/v3/recurrent_charge/livechat", payload{
		Name:              params.Name,
		Price:             params.Price,
		ReturnURL:         params.ReturnURL,
		Test:              params.Test,
		TrialDays:         params.TrialDays,
		Months:            params.Months,
		CommissionPercent: params.CommissionPercent,
	})
	if err != nil {
		return nil, err
	}

	return asRecurrentCharge(resp)
}

func (a *Api) GetRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error) {
	resp, err := a.call(ctx, "GET", "/v3/recurrent_charge/livechat/"+id, nil)
	if err != nil {
		return nil, err
	}

	return asRecurrentCharge(resp)
}

func (a *Api) CancelRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error) {
	resp, err := a.call(ctx, "PUT", fmt.Sprintf("/v3/recurrent_charge/livechat/%s/cancel", id), nil)
	if err != nil {
		return nil, err
	}

	return asRecurrentCharge(resp)
}

func (a *Api) ActivateRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error) {
	resp, err := a.call(ctx, "PUT", fmt.Sprintf("/v3/recurrent_charge/livechat/%s/activate", id), nil)
	if err != nil {
		return nil, err
	}
	return asRecurrentCharge(resp)
}

func (a *Api) ActivateDirectCharge(ctx context.Context, id string) (*DirectCharge, error) {
	resp, err := a.call(ctx, "PUT", fmt.Sprintf("/v3/direct_charge/livechat/%s/activate", id), nil)
	if err != nil {
		return nil, err
	}
	return asDirectCharge(resp)
}

func asDirectCharge(body []byte) (*DirectCharge, error) {
	var dc DirectCharge
	if err := json.Unmarshal(body, &dc); err != nil {
		return nil, err
	}

	return &dc, nil
}

func asRecurrentCharge(body []byte) (*RecurrentCharge, error) {
	var rc RecurrentCharge
	if err := json.Unmarshal(body, &rc); err != nil {
		return nil, err
	}

	return &rc, nil
}

func (a *Api) call(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.ApiBaseURL+path, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("billing api: create request: %w", err)
	}

	token, err := a.TokenFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("billing api: get token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("billing api: empty token")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("biling api: send call: %w", err)
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		var e []byte
		if resp.Body != nil {
			e, _ = io.ReadAll(resp.Body)
		}

		return nil, fmt.Errorf("biling api call to %s: bad status code %d, response: %s", path, resp.StatusCode, string(e))
	}

	return io.ReadAll(resp.Body)
}
