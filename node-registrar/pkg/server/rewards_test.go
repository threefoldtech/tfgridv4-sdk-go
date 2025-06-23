package server

import (
	"math"
	"testing"
	"time"

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
		assert.Equal(t, Reward{
			UpTimePercentage: upTimePercentage,
		}, got)
		return
	}

	// Calculate rewards
	memoryReward := bytesToGB(resources.MRU) * MemoryRewardPerGB
	ssdReward := bytesToTB(resources.SRU) * SsdRewardPerTB
	hddReward := bytesToTB(resources.HRU) * HddRewardPerTB

	// Calculate total rewards
	total := memoryReward + ssdReward + hddReward

	// Apply uptime percentage
	total = total * (upTimePercentage / 100)

	expected := Reward{
		FarmerReward:     total * FarmerRewardPercentage,
		TfReward:         total * TfRewardPercentage,
		FpReward:         total * FpRewardPercentage,
		Total:            total,
		UpTimePercentage: upTimePercentage,
	}

	// Use precise floating point comparison
	const delta = 1e-9 // Very small acceptable difference
	assert.InDelta(t, expected.FarmerReward, got.FarmerReward, delta)
	assert.InDelta(t, expected.TfReward, got.TfReward, delta)
	assert.InDelta(t, expected.FpReward, got.FpReward, delta)
	assert.InDelta(t, expected.Total, got.Total, delta)
	assert.InDelta(t, expected.UpTimePercentage, got.UpTimePercentage, delta)
}

// TestCalculatePeriodStart tests the calculatePeriodStart function with different inputs
func TestCalculatePeriodStart(t *testing.T) {
	tests := []struct {
		name         string
		inputTime    time.Time
		expectedTime time.Time
	}{
		{
			name:         "First period start timestamp",
			inputTime:    time.Unix(FirstPeriodStartTimestamp, 0),
			expectedTime: time.Unix(FirstPeriodStartTimestamp, 0),
		},
		{
			name:         "Exactly at second period start",
			inputTime:    time.Unix(FirstPeriodStartTimestamp+StandardPeriodDuration, 0),
			expectedTime: time.Unix(FirstPeriodStartTimestamp+StandardPeriodDuration, 0),
		},
		{
			name:         "Middle of a period",
			inputTime:    time.Unix(FirstPeriodStartTimestamp+StandardPeriodDuration/2, 0),
			expectedTime: time.Unix(FirstPeriodStartTimestamp, 0),
		},
		{
			name:         "Near end of a period",
			inputTime:    time.Unix(FirstPeriodStartTimestamp+StandardPeriodDuration-1, 0),
			expectedTime: time.Unix(FirstPeriodStartTimestamp, 0),
		},
		{
			name:         "Multiple periods later",
			inputTime:    time.Unix(FirstPeriodStartTimestamp+3*StandardPeriodDuration+StandardPeriodDuration/3, 0),
			expectedTime: time.Unix(FirstPeriodStartTimestamp+3*StandardPeriodDuration, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the actual function with our test time
			result := calculatePeriodStart(tt.inputTime)

			// Check that the result equals our expected time
			assert.Equal(t, tt.expectedTime.Unix(), result.Unix())

			// Additionally, verify that we can manually calculate the same result
			// secondsSinceCurrentPeriodStart := (tt.inputTime.Unix() - FirstPeriodStartTimestamp) % StandardPeriodDuration
			// manualCalculation := time.Unix(FirstPeriodStartTimestamp+secondsSinceCurrentPeriodStart, 0)
			// assert.Equal(t, manualCalculation.Unix(), result.Unix())
		})
	}
}

func TestCalculateUpTimePercentage(t *testing.T) {
	type args struct {
		reports     []db.UptimeReport
		periodStart time.Time
	}
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name      string
		args      args
		expected  float64
		wantError bool
	}{
		{
			name: "All uptime, no downtime (40 min gaps)",
			args: args{
				periodStart: now.Add(-160 * time.Minute), // Start 160 min ago (for 4 reports)
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-120 * time.Minute), Duration: 2400}, // 40 min (2400 seconds)
					{Timestamp: now.Add(-80 * time.Minute), Duration: 2400},  // 40 min (2400 seconds)
					{Timestamp: now.Add(-40 * time.Minute), Duration: 2400},  // 40 min (2400 seconds)
					{Timestamp: now, Duration: 2400},                         // 40 min (2400 seconds)
				},
			},
			expected: 100.0,
		},
		{
			name: "50% uptime — only half the reports received",
			args: args{
				periodStart: now.Add(-160 * time.Minute), // full 160 mins = 9600s
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-120 * time.Minute), Duration: 2400}, // 40 min (2400 seconds)
					{Timestamp: now.Add(-80 * time.Minute), Duration: 2400},  // 40 min (2400 seconds)
				},
			},
			expected: 50.0,
		},
		{
			name: "0% uptime — no reports received",
			args: args{
				periodStart: now.Add(-160 * time.Minute), // full 160 mins = 9600s
				reports:     []db.UptimeReport{},
			},
			expected: 0.0,
		},
		{
			name: "allowance for only one report received, after 1hour",
			args: args{
				periodStart: now.Add(-60 * time.Minute), // full 60 mins
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-40 * time.Minute), Duration: 2400}, // 40 min (2400 seconds)
				},
			},
			expected: 100.0,
		},
		{
			name: "one report received after 3 report intervals, with full duration of 120 min",
			args: args{
				periodStart: now.Add(-130 * time.Minute), // full 130 mins = 7800s
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-10 * time.Minute), Duration: 7200}, // 120 min (7200 seconds)
				},
			},
			expected: 100.0,
		},
		{
			name: "one report after single report interval with 30min uptime",
			args: args{
				periodStart: now.Add(-40 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now, Duration: 1200}, // 20 min (1200 seconds)
				},
			},
			expected: 50.0,
		},
		{
			name: "Duration decreases with time",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-80 * time.Minute), Duration: 2400}, // 40 min (2400 seconds)
					{Timestamp: now.Add(-40 * time.Minute), Duration: 1800}, // 30 min (1800 seconds)
					{Timestamp: now, Duration: 1200},                        // 20 min (1200 seconds)
				},
			},
			expected: 75.0,
		},
		{
			name: "Duration increases with time",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-80 * time.Minute), Duration: 1200}, // 20 min (1200 seconds)
					{Timestamp: now.Add(-40 * time.Minute), Duration: 1800}, // 30 min (1800 seconds)
					{Timestamp: now, Duration: 2400},                        // 40 min (2400 seconds)
				},
			},
			expected: 75.0,
		},
		{
			name: "Unordered reports - timestamps not in ascending order",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-40 * time.Minute), Duration: 2400}, // Out of order (should be before the one below)
					{Timestamp: now.Add(-80 * time.Minute), Duration: 2400}, 
					{Timestamp: now, Duration: 2400},
				},
			},
			expected: 0.0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateUpTimePercentage(tt.args.reports, tt.args.periodStart, now)

			// Check for expected error
			if tt.wantError {
				if err == nil {
					t.Errorf("calculateUpTimePercentage() expected error, got nil")
				}
				return
			}

			// No error expected
			if err != nil {
				t.Errorf("calculateUpTimePercentage() unexpected error: %v", err)
				return
			}
			if math.Abs(got-tt.expected) > 0.01 {
				t.Errorf("calculateUpTimePercentage() = %v, want %v", got, tt.expected)
			}
		})
	}
}
