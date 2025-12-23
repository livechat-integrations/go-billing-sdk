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

func TestSubscription_IsTrialActive_EndOfDay(t *testing.T) {
	now := time.Now()
	// Trial ended earlier today (e.g., 1 hour ago) - should still be active until end of day
	earlierToday := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 0, 0, 0, now.Location())

	tests := []struct {
		name         string
		subscription Subscription
		expected     bool
	}{
		{
			name: "trial ended earlier today - still active until end of day",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						TrialEndsAt:     &earlierToday,
						CurrentChargeAt: nil,
					}),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subscription.IsTrialActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscription_IsActive_EndOfDay(t *testing.T) {
	now := time.Now()
	// Trial ended earlier today - should still be active until end of day
	earlierToday := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, 0, 0, 0, now.Location())
	nextCharge := now.Add(30 * 24 * time.Hour)

	tests := []struct {
		name         string
		subscription Subscription
		expected     bool
	}{
		{
			name: "trial ended earlier today - still active until end of day",
			subscription: Subscription{
				Charge: &Charge{
					Payload: mustMarshal(livechat.RecurrentCharge{
						BaseCharge: livechat.BaseCharge{
							Status: "pending",
						},
						TrialEndsAt:  &earlierToday,
						NextChargeAt: &nextCharge,
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

func TestEndOfDay(t *testing.T) {
	// Test with a specific time
	loc, _ := time.LoadLocation("America/New_York")
	input := time.Date(2024, 6, 15, 10, 30, 45, 123456789, loc)

	result := endOfDay(input)

	assert.Equal(t, 2024, result.Year())
	assert.Equal(t, time.June, result.Month())
	assert.Equal(t, 15, result.Day())
	assert.Equal(t, 23, result.Hour())
	assert.Equal(t, 59, result.Minute())
	assert.Equal(t, 59, result.Second())
	assert.Equal(t, 999999999, result.Nanosecond())
	assert.Equal(t, loc, result.Location())
}

// Helper function to marshal JSON for tests
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
