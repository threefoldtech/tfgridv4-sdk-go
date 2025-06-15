package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

func TestCalculateMonthlyReward(t *testing.T) {
	// Define standard capacity for most tests
	standardCapacity := db.Resources{
		CRU: 8,
		MRU: 68719476736,
		SRU: 4398046511104,
		HRU: 17592186044416,
	}

	// Define a small capacity for precision testing
	preciseCapacity := db.Resources{
		CRU: 1,
		MRU: 1073741824, // 1 GB exactly
		SRU: 0,
		HRU: 0,
	}

	tests := []struct {
		name             string
		capacity         db.Resources
		upTimePercentage float64
		wantError        bool
		expectedError    error
	}{
		{
			name:             "negative uptime percentage",
			capacity:         standardCapacity,
			upTimePercentage: -1,
			wantError:        true,
			expectedError:    ErrInvalidUptimePercentage,
		},
		{
			name:             "valid uptime (5%)",
			capacity:         standardCapacity,
			upTimePercentage: 5,
			wantError:        false,
		},
		{
			name:             "uptime over 100%",
			capacity:         standardCapacity,
			upTimePercentage: 101,
			wantError:        true,
			expectedError:    ErrInvalidUptimePercentage,
		},
		{
			name:             "uptime at 100%",
			capacity:         standardCapacity,
			upTimePercentage: 100,
			wantError:        false,
		},
		{
			name:             "uptime below threshold (80%)",
			capacity:         standardCapacity,
			upTimePercentage: 80,
			wantError:        false,
		},
		{
			name:             "uptime at threshold (90%)",
			capacity:         standardCapacity,
			upTimePercentage: 90,
			wantError:        false,
		},
		{
			name:             "uptime below threshold as float (89.1%)",
			capacity:         standardCapacity,
			upTimePercentage: 89.1,
			wantError:        false,
		},
		{
			name:             "uptime at 98%",
			capacity:         standardCapacity,
			upTimePercentage: 98,
			wantError:        false,
		},
		{
			name:             "precision test with small capacity",
			capacity:         preciseCapacity,
			upTimePercentage: 100,
			wantError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateMonthlyReward(tt.capacity, tt.upTimePercentage)

			// Error check
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.Equal(t, tt.expectedError, err)
				return
			}

			// No error expected
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check reward calculation
			AssertMonthlyReward(t, tt.capacity, tt.upTimePercentage, got)
		})
	}
}

func AssertMonthlyReward(t testing.TB, resources db.Resources, upTimePercentage float64, got Reward) {
	t.Helper()

	if upTimePercentage < 90 {
		assert.Equal(t, Reward{}, got)
		return
	}

	// Calculate rewards
	memoryReward := bytesToGB(resources.MRU) * MEMORY_REWARD_PER_GB
	ssdReward := bytesToTB(resources.SRU) * SSD_REWARD_PER_TB
	hddReward := bytesToTB(resources.HRU) * HDD_REWARD_PER_TB

	// Calculate total rewards
	total := memoryReward + ssdReward + hddReward

	// Apply uptime percentage
	total = total * (upTimePercentage / 100)

	expected := Reward{
		FarmerReward: total * FARMER_REWARD_PERCENTAGE,
		TFReward:     total * TF_REWARD_PERCENTAGE,
		FPReward:     total * FP_REWARD_PERCENTAGE,
		Total:        total,
	}

	// Use precise floating point comparison
	const delta = 1e-9 // Very small acceptable difference
	assert.InDelta(t, expected.FarmerReward, got.FarmerReward, delta)
	assert.InDelta(t, expected.TFReward, got.TFReward, delta)
	assert.InDelta(t, expected.FPReward, got.FPReward, delta)
	assert.InDelta(t, expected.Total, got.Total, delta)
}
