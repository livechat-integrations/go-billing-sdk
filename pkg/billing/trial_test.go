package billing

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/livechat-integrations/go-billing-sdk/v2/internal/livechat"
	"github.com/stretchr/testify/assert"
)

func TestSubscription_IsTrialActive(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(3 * 24 * time.Hour) // 3 days in future
	pastDate := now.Add(-2 * 24 * time.Hour)  // 2 days ago

	tests := []struct {
		name         string
		subscription Subscription
		expected     bool
	}{
		{
			name: "active trial - trial ends in future and no current charge",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						TrialEndsAt:     &futureDate,
						CurrentChargeAt: nil,
					}),
				},
			},
			expected: true,
		},
		{
			name: "expired trial - trial ended in past",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						TrialEndsAt:     &pastDate,
						CurrentChargeAt: nil,
					}),
				},
			},
			expected: false,
		},
		{
			name: "converted trial - has current charge",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						TrialEndsAt:     &futureDate,
						CurrentChargeAt: &now,
					}),
				},
			},
			expected: false,
		},
		{
			name: "no trial - no trial ends at",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						CurrentChargeAt: nil,
					}),
				},
			},
			expected: false,
		},
		{
			name:         "no charge",
			subscription: Subscription{},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subscription.IsTrialActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscription_IsActive_WithTrial(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(3 * 24 * time.Hour)  // 3 days in future
	pastDate := now.Add(-2 * 24 * time.Hour)   // 2 days ago
	nextCharge := now.Add(30 * 24 * time.Hour) // 30 days in future

	tests := []struct {
		name         string
		subscription Subscription
		expected     bool
	}{
		{
			name: "active trial subscription",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						BaseCharge: livechat.BaseCharge{
							Status: "pending",
						},
						TrialEndsAt:  &futureDate,
						NextChargeAt: &nextCharge,
					}),
				},
			},
			expected: true,
		},
		{
			name: "expired trial without payment",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						BaseCharge: livechat.BaseCharge{
							Status: "pending",
						},
						TrialEndsAt:  &pastDate,
						NextChargeAt: &nextCharge,
					}),
				},
			},
			expected: false,
		},
		{
			name: "trial converted to paid subscription",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						BaseCharge: livechat.BaseCharge{
							Status: "active",
						},
						TrialEndsAt:     &pastDate,
						CurrentChargeAt: &now,
						NextChargeAt:    &nextCharge,
					}),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subscription.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to marshal JSON for tests
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
