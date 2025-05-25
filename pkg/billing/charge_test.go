package billing

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPlans_GetPlan(t *testing.T) {
	plans := Plans{
		{
			Name: "plan1",
		},
		{
			Name: "plan2",
		},
	}

	t.Run("plan found", func(t *testing.T) {
		assert.NotNil(t, plans.GetPlan("plan1"))
	})

	t.Run("plan not found", func(t *testing.T) {
		assert.Nil(t, plans.GetPlan("plan3"))
	})
}

func TestSubscription_IsActive(t *testing.T) {
	t.Run("charge is nil", func(t *testing.T) {
		subscription := Subscription{
			Charge: nil,
		}
		assert.True(t, subscription.IsActive())
	})

	t.Run("charge is active", func(t *testing.T) {
		subscription := Subscription{
			Charge: &Charge{
				Payload: []byte(`{"status": "active", "next_charge_at": "` + time.Now().AddDate(0, 0, 1).Format(time.RFC3339) + `"}`),
			},
		}
		assert.True(t, subscription.IsActive())
	})

	t.Run("charge is canceled", func(t *testing.T) {
		subscription := Subscription{
			Charge: &Charge{
				CanceledAt: nil,
				Payload:    []byte(`{"status": "canceled"}`),
			},
		}
		assert.False(t, subscription.IsActive())
	})

	t.Run("charge is past_due, but in the retention period", func(t *testing.T) {
		subscription := Subscription{
			Charge: &Charge{
				Payload: []byte(`{"status": "past_due", "next_charge_at": "` + time.Now().AddDate(0, 0, -1).Format(time.RFC3339) + `"}`),
			},
		}

		assert.True(t, subscription.IsActive())
	})

	t.Run("charge is past_due, after retention period", func(t *testing.T) {
		subscription := Subscription{
			Charge: &Charge{
				Payload: []byte(`{"status": "past_due", "next_charge_at": "` + time.Now().AddDate(0, 0, -5).Format(time.RFC3339) + `"}`),
			},
		}

		assert.False(t, subscription.IsActive())
	})

	t.Run("charge is active, next_charge_at null", func(t *testing.T) {
		subscription := Subscription{
			Charge: &Charge{
				Payload: []byte(`{"status": "active", "next_charge_at": null}`),
			},
		}

		assert.False(t, subscription.IsActive())
	})
}
