package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const billingAPIBaseURL = "https://billing.livechatinc.com"

type BaseCharge struct {
	ID                string     `json:"id"`
	BuyerLicenseID    int        `json:"buyer_license_id"`
	BuyerEntityID     string     `json:"buyer_entity_id"`
	SellerClientID    string     `json:"seller_client_id"`
	OrderClientID     string     `json:"order_client_id"`
	OrderLicenseID    string     `json:"order_license_id"`
	OrderEntityID     string     `json:"order_entity_id"`
	Name              string     `json:"name"`
	Price             int        `json:"price"`
	ReturnURL         string     `json:"return_url"`
	Test              bool       `json:"test"`
	PerAccount        bool       `json:"per_account"`
	Status            string     `json:"status"`
	ConfirmationURL   string     `json:"confirmation_url"`
	CommissionPercent int        `json:"commission_percent"`
	UpdatedAt         *time.Time `json:"updated_at"`
	CreatedAt         *time.Time `json:"created_at"`
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

type apiInterface interface {
	CreateRecurrentCharge(ctx context.Context, params createRecurrentChargeParams) (*RecurrentCharge, error)
	GetRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error)
}

type httpCaller interface {
	Do(req *http.Request) (*http.Response, error)
}

type TokenFn func(ctx context.Context) (string, error)

type api struct {
	httpClient httpCaller
	apiBaseURL string
	tokenFn    TokenFn
}

type createRecurrentChargeParams struct {
	Name              string
	ReturnURL         string
	Price             int
	Test              bool
	TrialDays         int
	Months            int
	CommissionPercent *int
}

func (a *api) CreateRecurrentCharge(ctx context.Context, params createRecurrentChargeParams) (*RecurrentCharge, error) {
	type payload struct {
		Name              string `json:"name"`
		Price             int    `json:"price"`
		ReturnURL         string `json:"return_url"`
		Test              bool   `json:"test"`
		TrialDays         int    `json:"trial_days"`
		Months            int    `json:"months,omitempty"`
		CommissionPercent *int   `json:"commission_percent,omitempty"`
	}
	resp, err := a.call(ctx, "POST", "/v1/recurrent_charge", payload{
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

func (a *api) GetRecurrentCharge(ctx context.Context, id string) (*RecurrentCharge, error) {
	resp, err := a.call(ctx, "GET", "/v1/recurrent_charge/"+id, nil)
	if err != nil {
		return nil, err
	}

	return asRecurrentCharge(resp)
}

func asRecurrentCharge(body []byte) (*RecurrentCharge, error) {
	var rc RecurrentCharge
	if err := json.Unmarshal(body, &rc); err != nil {
		return nil, err
	}

	return &rc, nil
}

func (a *api) call(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.apiBaseURL+path, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("billing api: create request: %w", err)
	}

	token, err := a.tokenFn(ctx)
	if err != nil {
		return nil, fmt.Errorf("billing api: get token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("billing api: empty token")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
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
