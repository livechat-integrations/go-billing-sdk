package livechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var hm = new(httpMock)

var a = Api{
	HttpClient: hm,
	ApiBaseURL: "https://api.livechatinc.com",
	TokenFn: func(ctx context.Context) (string, error) {
		return "token", nil
	},
}

type httpMock struct {
	mock.Mock
}

func (m *httpMock) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestAPI_CreateRecurrentCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(strings.NewReader(`{"id":"1"}`)),
		}, nil).Once()

		charge, err := a.CreateRecurrentCharge(context.Background(), CreateRecurrentChargeParams{})
		assert.NoError(t, err)
		assert.Equal(t, "1", charge.ID)
	})

	t.Run("error", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(``)),
		}, nil).Once()

		charge, err := a.CreateRecurrentCharge(context.Background(), CreateRecurrentChargeParams{})
		assert.Error(t, err)
		assert.Nil(t, charge)
	})
}

func TestAPI_GetRecurrentCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"id":"1"}`)),
		}, nil).Once()

		charge, err := a.GetRecurrentCharge(context.Background(), "1")
		assert.NoError(t, err)
		assert.Equal(t, "1", charge.ID)
	})

	t.Run("error", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(``)),
		}, nil).Once()

		charge, err := a.GetRecurrentCharge(context.Background(), "1")
		assert.Error(t, err)
		assert.Nil(t, charge)
	})
}

func TestAPI_call(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"id":"1"}`)),
		}, nil).Once()

		resp, err := a.call(context.Background(), "GET", "/", nil)
		var r struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(resp, &r)
		assert.NoError(t, err)
		assert.Equal(t, "1", r.ID)
	})

	t.Run("error create request", func(t *testing.T) {
		_, err := a.call(nil, "GET", "/", nil) //nolint:staticcheck
		assert.ErrorContains(t, err, "create request")
	})

	t.Run("empty token", func(t *testing.T) {
		a.TokenFn = func(ctx context.Context) (string, error) {
			return "", nil
		}
		_, err := a.call(context.Background(), "GET", "/", nil)
		assert.ErrorContains(t, err, "empty token")
	})

	t.Run("error get token", func(t *testing.T) {
		a.TokenFn = func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("error")
		}
		_, err := a.call(context.Background(), "GET", "/", nil)
		assert.ErrorContains(t, err, "get token")
	})

	t.Run("error send request", func(t *testing.T) {
		a.TokenFn = func(ctx context.Context) (string, error) {
			return "token", nil
		}
		hm.On("Do", mock.Anything).Return(&http.Response{}, fmt.Errorf("error")).Once()

		resp, err := a.call(context.Background(), "GET", "/", nil)
		assert.ErrorContains(t, err, "send call")
		assert.Nil(t, resp)
	})

	t.Run("404", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 404,
		}, nil).Once()

		resp, err := a.call(context.Background(), "GET", "/", nil)
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("500", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 500,
		}, nil).Once()

		resp, err := a.call(context.Background(), "GET", "/", nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("500 with body", func(t *testing.T) {
		hm.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(`{"error":"error"}`)),
		}, nil).Once()

		resp, err := a.call(context.Background(), "GET", "/", nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func Test_asRecurrentCharge(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rc, err := asRecurrentCharge([]byte(`{"id":"1"}`))
		assert.NoError(t, err)
		assert.Equal(t, "1", rc.ID)
	})

	t.Run("error", func(t *testing.T) {
		rc, err := asRecurrentCharge([]byte(``))
		assert.Error(t, err)
		assert.Nil(t, rc)
	})
}
