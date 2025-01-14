package billing

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
				Payload: []byte(`{"status": "active"}`),
			},
		}
		assert.True(t, subscription.IsActive())
	})

	t.Run("charge is success", func(t *testing.T) {
		subscription := Subscription{
			Charge: &Charge{
				Payload: []byte(`{"status": "success"}`),
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
}
