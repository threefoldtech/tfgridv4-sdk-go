package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

func TestComputeCapacityRewardWithUptime(t *testing.T) {
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
			name:             fmt.Sprintf("uptime at threshold (%.0f%%)", MinUptimePercentageForReward),
			capacity:         standardCapacity,
			upTimePercentage: MinUptimePercentageForReward,
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
			got, err := computeCapacityRewardWithUptime(tt.capacity, tt.upTimePercentage)

			// Error check
			if tt.wantError {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err)
				return
			}

			// No error expected
			require.NoError(t, err)

			// Check reward calculation
			AssertCapacityReward(t, tt.capacity, tt.upTimePercentage, got)
		})
	}
}

func AssertCapacityReward(t testing.TB, resources db.Resources, upTimePercentage float64, got Reward) {
	t.Helper()

	if upTimePercentage < MinUptimePercentageForReward {
		assert.Equal(t, Reward{UpTimePercentage: upTimePercentage}, got)
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

	// Apply truncation to match the implementation
	expected := Reward{
		FarmerReward:     truncateFloat(total*FarmerRewardPercentage, RewardPrecisionDecimalPlaces),
		TfReward:         truncateFloat(total*TfRewardPercentage, RewardPrecisionDecimalPlaces),
		FpReward:         truncateFloat(total*FpRewardPercentage, RewardPrecisionDecimalPlaces),
		Total:            truncateFloat(total, RewardPrecisionDecimalPlaces),
		UpTimePercentage: upTimePercentage,
	}

	assert.Equal(t, expected, got)

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

func TestCalculateTotalReward(t *testing.T) {
	tests := []struct {
		name             string
		capacity         db.Resources
		upTimePercentage float64
		expected         float64
	}{
		{
			name: "standard capacity with 100% uptime",
			capacity: db.Resources{
				CRU: 8,
				MRU: 68719476736,    // 64 GB
				SRU: 1099511627776,  // 1 TB
				HRU: 10995116277760, // 10 TB
			},
			upTimePercentage: 100,
			expected:         512 + 31.5 + 70, // (64 * 8) + (1 * 31.5) + (10 * 7)
		},
		{
			name: "standard capacity with 50% uptime",
			capacity: db.Resources{
				CRU: 8,
				MRU: 68719476736,    // 64 GB
				SRU: 1099511627776,  // 1 TB
				HRU: 10995116277760, // 10 TB
			},
			upTimePercentage: 50,
			expected:         (512 + 31.5 + 70) * 0.5, // 50% of total
		},
		{
			name: "zero capacity with 100% uptime",
			capacity: db.Resources{
				CRU: 0,
				MRU: 0,
				SRU: 0,
				HRU: 0,
			},
			upTimePercentage: 100,
			expected:         0,
		},
		{
			name: "small memory only capacity with 95% uptime",
			capacity: db.Resources{
				CRU: 0,
				MRU: 1073741824, // 1 GB
				SRU: 0,
				HRU: 0,
			},
			upTimePercentage: 95,
			expected:         8 * 0.95, // (1 * 8) * 0.95
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTotalReward(tt.capacity, tt.upTimePercentage)

			truncatedExpected := truncateFloat(tt.expected, RewardPrecisionDecimalPlaces)
			assert.Equal(t, truncatedExpected, result, "Total reward calculation incorrect")
		})
	}
}

// TestCalculateBaseCapacityReward tests the calculateBaseCapacityReward function
func TestCalculateBaseCapacityReward(t *testing.T) {
	tests := []struct {
		name     string
		capacity db.Resources
		expected float64
	}{
		{
			name: "standard capacity",
			capacity: db.Resources{
				CRU: 8,
				MRU: 68719476736,    // 64 GB
				SRU: 1099511627776,  // 1 TB
				HRU: 10995116277760, // 10 TB
			},
			expected: 512 + 31.5 + 70, // (64 * 8) + (1 * 31.5) + (10 * 7)
		},
		{
			name: "zero capacity",
			capacity: db.Resources{
				CRU: 0,
				MRU: 0,
				SRU: 0,
				HRU: 0,
			},
			expected: 0,
		},
		{
			name: "memory only",
			capacity: db.Resources{
				CRU: 0,
				MRU: 1073741824, // 1 GB
				SRU: 0,
				HRU: 0,
			},
			expected: 8, // 1 * 8
		},
		{
			name: "SSD only",
			capacity: db.Resources{
				CRU: 0,
				MRU: 0,
				SRU: 1099511627776, // 1 TB
				HRU: 0,
			},
			expected: 31.5, // 1 * 31.5
		},
		{
			name: "HDD only",
			capacity: db.Resources{
				CRU: 0,
				MRU: 0,
				SRU: 0,
				HRU: 1099511627776, // 1 TB
			},
			expected: 7, // 1 * 7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBaseCapacityReward(tt.capacity)

			truncatedExpected := truncateFloat(tt.expected, RewardPrecisionDecimalPlaces)
			assert.Equal(t, truncatedExpected, result, "Base capacity reward calculation incorrect")
		})
	}
}

// TestRewardFloatingPointPrecision tests that the floating point calculations in CalculateCapacityReward
// are handled correctly and the percentages are applied as expected
func TestRewardFloatingPointPrecision(t *testing.T) {
	// Create a resource with a specific memory size to test floating point precision
	memoryOnlyCapacity := db.Resources{
		CRU: 0,
		MRU: 1073741824, // 1 GB exactly
		SRU: 0,
		HRU: 0,
	}

	// Calculate expected values based on our constants
	expectedBaseReward := bytesToGB(memoryOnlyCapacity.MRU) * MemoryRewardPerGB // Should be 8.0
	expectedTotal := expectedBaseReward                                         // 100% uptime

	// Expected distribution based on percentages
	expectedFarmerReward := expectedTotal * FarmerRewardPercentage // 8.0 * 0.6 = 4.8
	expectedTfReward := expectedTotal * TfRewardPercentage         // 8.0 * 0.2 = 1.6
	expectedFpReward := expectedTotal * FpRewardPercentage         // 8.0 * 0.2 = 1.6

	// Get the actual reward calculation
	reward, err := computeCapacityRewardWithUptime(memoryOnlyCapacity, 100)
	require.NoError(t, err)

	// Test precision and distribution
	t.Run("base reward calculation", func(t *testing.T) {
		expectedTruncated := truncateFloat(expectedBaseReward, RewardPrecisionDecimalPlaces)
		actualResult := calculateBaseCapacityReward(memoryOnlyCapacity)
		assert.Equal(t, expectedTruncated, actualResult)
	})

	t.Run("total reward", func(t *testing.T) {
		assert.Equal(t, truncateFloat(expectedTotal, RewardPrecisionDecimalPlaces), reward.Total)
	})

	// Test the distribution percentages
	t.Run("farmer reward percentage", func(t *testing.T) {
		assert.Equal(t, truncateFloat(expectedFarmerReward, RewardPrecisionDecimalPlaces), reward.FarmerReward)
		ratio := truncateFloat(reward.FarmerReward/reward.Total, 3)
		assert.Equal(t, 0.6, ratio, "Farmer reward should be 60% of total")
	})

	t.Run("tf reward percentage", func(t *testing.T) {
		assert.Equal(t, truncateFloat(expectedTfReward, RewardPrecisionDecimalPlaces), reward.TfReward)
		ratio := truncateFloat(reward.TfReward/reward.Total, 3)
		assert.Equal(t, 0.2, ratio, "TF reward should be 20% of total")
	})

	t.Run("fp reward percentage", func(t *testing.T) {
		assert.Equal(t, truncateFloat(expectedFpReward, RewardPrecisionDecimalPlaces), reward.FpReward)
		ratio := truncateFloat(reward.FpReward/reward.Total, 3)
		assert.Equal(t, 0.2, ratio, "FP reward should be 20% of total")
	})

	// Test that sum of portions equals total (within rounding error)
	t.Run("reward portions sum to total", func(t *testing.T) {
		actualSum := reward.FarmerReward + reward.TfReward + reward.FpReward
		totalTruncated := truncateFloat(reward.Total, RewardPrecisionDecimalPlaces)
		sumTruncated := truncateFloat(actualSum, RewardPrecisionDecimalPlaces)
		assert.Equal(t, totalTruncated, sumTruncated, "Sum of reward portions should equal total reward")
	})
}

// TestAreReportsOrderedCorrectly tests the areReportsOrderedCorrectly function
func TestCalculateCapacityRewardd(t *testing.T) {
	now := time.Now()
	periodStart := now.Add(-24 * time.Hour)
	periodEnd := now

	// Define standard capacity for tests
	standardCapacity := db.Resources{
		CRU: 8,
		MRU: 68719476736,    // 64 GB
		SRU: 1099511627776,  // 1 TB
		HRU: 10995116277760, // 10 TB
	}

	t.Run("Test ErrNoReportsAvailable", func(t *testing.T) {
		_, err := CalculateCapacityReward(standardCapacity, []db.UptimeReport{}, periodStart, periodEnd)
		assert.ErrorIs(t, err, ErrNoReportsAvailable)
	})
	t.Run("Test ErrReportsNotInAscOrder", func(t *testing.T) {
		_, err := CalculateCapacityReward(standardCapacity, []db.UptimeReport{
			{
				Timestamp: now.Add(-2 * time.Hour),
				Duration:  2 * time.Hour,
			},
			{
				Timestamp: now.Add(-6 * time.Hour),
				Duration:  1 * time.Hour,
			},
		}, periodStart, periodEnd)
		assert.ErrorIs(t, err, ErrReportsNotInAscOrder)
	})
}

func TestAreReportsOrderedCorrectly(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		reports  []db.UptimeReport
		expected bool
	}{
		{
			name:     "empty reports",
			reports:  []db.UptimeReport{},
			expected: true, // empty reports are considered properly ordered
		},
		{
			name: "single report",
			reports: []db.UptimeReport{
				{Timestamp: now},
			},
			expected: true, // single report is always ordered
		},
		{
			name: "ordered reports",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-2 * time.Hour)},
				{Timestamp: now.Add(-1 * time.Hour)},
				{Timestamp: now},
			},
			expected: true,
		},
		{
			name: "unordered reports",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-1 * time.Hour)},
				{Timestamp: now.Add(-2 * time.Hour)}, // out of order
				{Timestamp: now},
			},
			expected: false,
		},
		{
			name: "same timestamps",
			reports: []db.UptimeReport{
				{Timestamp: now},
				{Timestamp: now}, // same timestamp
			},
			expected: true, // equal timestamps are considered ordered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := areReportsOrderedCorrectly(tt.reports)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCalculateDowntimeFromReports tests the calculateDowntimeFromReports function
func TestCalculateDowntimeFromReports(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name     string
		reports  []db.UptimeReport
		expected time.Duration
	}{
		{
			name:     "empty reports",
			reports:  []db.UptimeReport{},
			expected: 0,
		},
		{
			name: "single report",
			reports: []db.UptimeReport{
				{Timestamp: now, Duration: time.Hour},
			},
			expected: 0, // single report means no gaps to calculate
		},
		{
			name: "no downtime - perfect reports",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-2 * time.Hour), Duration: time.Hour},
				{Timestamp: now.Add(-1 * time.Hour), Duration: time.Hour},
			},
			expected: 0, // no downtime
		},
		{
			name: "partial downtime - gap larger than duration",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-3 * time.Hour), Duration: time.Hour},
				{Timestamp: now.Add(-1 * time.Hour), Duration: time.Hour}, // 2 hour gap, 1 hour reported
			},
			expected: time.Hour, // 1 hour of downtime
		},
		{
			name: "decreasing duration but no downtime due to equal gap and duration",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-2 * time.Hour), Duration: 2 * time.Hour},
				{Timestamp: now.Add(-1 * time.Hour), Duration: time.Hour}, // duration decreased but gap == duration
			},
			expected: 0, // No downtime since gap == duration
		},
		{
			name: "decreasing duration with actual downtime",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-3 * time.Hour), Duration: 2 * time.Hour},
				{Timestamp: now.Add(-1 * time.Hour), Duration: time.Hour}, // duration decreased and gap > duration
			},
			expected: time.Hour, // 1 hour of downtime (2 hour gap - 1 hour duration)
		},
		{
			name: "multiple downtime periods",
			reports: []db.UptimeReport{
				{Timestamp: now.Add(-4 * time.Hour), Duration: time.Hour},
				{Timestamp: now.Add(-3 * time.Hour), Duration: 30 * time.Minute}, // 30 min downtime
				{Timestamp: now.Add(-1 * time.Hour), Duration: 30 * time.Minute}, // 90 min downtime (2h - 30min)
			},
			expected: 30*time.Minute + 90*time.Minute, // total 2 hours of downtime
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDowntimeFromReports(tt.reports)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCalculatePercentage tests the calculatePercentage function
func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		name        string
		totalPeriod time.Duration
		downtime    time.Duration
		expected    float64
	}{
		{
			name:        "no downtime",
			totalPeriod: 24 * time.Hour,
			downtime:    0,
			expected:    100.0,
		},
		{
			name:        "50% downtime",
			totalPeriod: 24 * time.Hour,
			downtime:    12 * time.Hour,
			expected:    50.0,
		},
		{
			name:        "total downtime",
			totalPeriod: 24 * time.Hour,
			downtime:    24 * time.Hour,
			expected:    0.0,
		},
		{
			name:        "25% downtime",
			totalPeriod: 24 * time.Hour,
			downtime:    6 * time.Hour,
			expected:    75.0,
		},
		{
			name:        "small percentage downtime",
			totalPeriod: 1000 * time.Hour,
			downtime:    1 * time.Hour,
			expected:    99.9, // truncated to 2 decimal places
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercentage(tt.totalPeriod, tt.downtime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDowntimeSinceLastReportTimestamp tests the downtimeSinceLastReportTimestamp function
func TestDowntimeSinceLastReportTimestamp(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name                string
		lastReportTimestamp time.Time
		currentTime         time.Time
		expected            time.Duration
	}{
		{
			name:                "recent report, no downtime",
			lastReportTimestamp: now.Add(-time.Duration(UptimeEventsInterval-100) * time.Second),
			currentTime:         now,
			expected:            0,
		},
		{
			name:                "exactly at threshold",
			lastReportTimestamp: now.Add(-time.Duration(UptimeEventsInterval) * time.Second),
			currentTime:         now,
			expected:            time.Duration(UptimeEventsInterval) * time.Second,
		},
		{
			name:                "over threshold",
			lastReportTimestamp: now.Add(-2 * time.Duration(UptimeEventsInterval) * time.Second),
			currentTime:         now,
			expected:            2 * time.Duration(UptimeEventsInterval) * time.Second,
		},
		{
			name:                "future timestamp, no downtime",
			lastReportTimestamp: now.Add(time.Hour),
			currentTime:         now,
			expected:            0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := downtimeSinceLastReportTimestamp(tt.lastReportTimestamp, tt.currentTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHelper functions tests the various helper functions used in reward calculations
func TestHelperFunctions(t *testing.T) {
	// Test bytesToGB
	t.Run("bytesToGB conversion", func(t *testing.T) {
		t.Run("zero bytes", func(t *testing.T) {
			result := bytesToGB(0)
			assert.Equal(t, 0.0, result, "0 bytes should convert to 0 GB")
		})

		t.Run("1 GB", func(t *testing.T) {
			result := bytesToGB(1073741824) // 1 GB in bytes (2^30)
			truncatedResult := truncateFloat(result, 3)
			assert.Equal(t, 1.0, truncatedResult, "1073741824 bytes should convert to 1 GB")
		})

		t.Run("1.5 GB", func(t *testing.T) {
			result := bytesToGB(1610612736) // 1.5 GB in bytes
			truncatedResult := truncateFloat(result, 3)
			assert.Equal(t, 1.5, truncatedResult, "1610612736 bytes should convert to 1.5 GB")
		})
	})

	// Test bytesToTB
	t.Run("bytesToTB conversion", func(t *testing.T) {
		t.Run("zero bytes", func(t *testing.T) {
			result := bytesToTB(0)
			assert.Equal(t, 0.0, result, "0 bytes should convert to 0 TB")
		})

		t.Run("1 TB", func(t *testing.T) {
			result := bytesToTB(1099511627776) // 1 TB in bytes (2^40)
			truncatedResult := truncateFloat(result, 3)
			assert.Equal(t, 1.0, truncatedResult, "1099511627776 bytes should convert to 1 TB")
		})

		t.Run("2.5 TB", func(t *testing.T) {
			result := bytesToTB(2748779069440) // 2.5 TB in bytes
			truncatedResult := truncateFloat(result, 3)
			assert.Equal(t, 2.5, truncatedResult, "2748779069440 bytes should convert to 2.5 TB")
		})
	})

	// Test truncateFloat
	t.Run("truncateFloat", func(t *testing.T) {
		t.Run("truncate to 2 places", func(t *testing.T) {
			result := truncateFloat(123.456789, 2)
			assert.Equal(t, 123.45, result, "123.456789 truncated to 2 decimal places should be 123.45")
		})

		t.Run("truncate 124", func(t *testing.T) {
			result := truncateFloat(124, 0)
			assert.Equal(t, 124.0, result, "124 truncated to 0 decimal places should be 124.0")
		})

		t.Run("truncate to 0 places", func(t *testing.T) {
			result := truncateFloat(123.456789, 0)
			assert.Equal(t, 123.0, result, "123.456789 truncated to 0 decimal places should be 123.0")
		})

		t.Run("truncate to 3 places", func(t *testing.T) {
			result := truncateFloat(123.456789, 3)
			assert.Equal(t, 123.456, result, "123.456789 truncated to 3 decimal places should be 123.456")
		})

		t.Run("truncate negative number", func(t *testing.T) {
			result := truncateFloat(-123.456789, 2)
			assert.Equal(t, -123.45, result, "-123.456789 truncated to 2 decimal places should be -123.45")
		})
	})
}

func TestCalculateUpTimePercentage(t *testing.T) {
	type args struct {
		reports     []db.UptimeReport
		periodStart time.Time
	}
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name          string
		args          args
		expected      float64
		wantError     bool
		expectedError error
	}{
		{
			name: "All uptime, no downtime (40 min gaps)",
			args: args{
				periodStart: now.Add(-160 * time.Minute), // Start 160 min ago (for 4 reports)
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-120 * time.Minute), Duration: 40 * time.Minute}, // 40 min
					{Timestamp: now.Add(-80 * time.Minute), Duration: 40 * time.Minute},  // 40 min
					{Timestamp: now.Add(-40 * time.Minute), Duration: 40 * time.Minute},  // 40 min
					{Timestamp: now, Duration: 40 * time.Minute},                         // 40 min
				},
			},
			expected: 100.0,
		},
		{
			name: "50% uptime — only half the reports received",
			args: args{
				periodStart: now.Add(-160 * time.Minute), // full 160 mins = 9600s
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-120 * time.Minute), Duration: 40 * time.Minute}, // 40 min
					{Timestamp: now.Add(-80 * time.Minute), Duration: 40 * time.Minute},  // 40 min
				},
			},
			expected: 50.0,
		},
		{
			name: "Empty reports — should return error",
			args: args{
				periodStart: now.Add(-160 * time.Minute), // full 160 mins = 9600s
				reports:     []db.UptimeReport{},
			},
			expected:      0.0,
			wantError:     true,
			expectedError: ErrNoReportsAvailable,
		},
		{
			name: "allowance for only one report received, after 1hour",
			args: args{
				periodStart: now.Add(-60 * time.Minute), // full 60 mins
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-40 * time.Minute), Duration: 40 * time.Minute}, // 40 min
				},
			},
			expected: 100.0,
		},
		{
			name: "one report received after 3 report intervals, with full duration of 120 min",
			args: args{
				periodStart: now.Add(-130 * time.Minute), // full 130 mins = 7800s
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-10 * time.Minute), Duration: 120 * time.Minute}, // 120 min
				},
			},
			expected: 100.0,
		},
		{
			name: "one report after single report interval with 30min uptime",
			args: args{
				periodStart: now.Add(-40 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now, Duration: 20 * time.Minute}, // 20 min
				},
			},
			expected: 50.0,
		},
		{
			name: "Duration decreases with time",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-80 * time.Minute), Duration: 40 * time.Minute}, // 40 min
					{Timestamp: now.Add(-40 * time.Minute), Duration: 30 * time.Minute}, // 30 min
					{Timestamp: now, Duration: 20 * time.Minute},                        // 20 min
				},
			},
			expected: 75.0,
		},
		{
			name: "Duration increases with time",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-80 * time.Minute), Duration: 20 * time.Minute}, // 20 min
					{Timestamp: now.Add(-40 * time.Minute), Duration: 30 * time.Minute}, // 30 min
					{Timestamp: now, Duration: 40 * time.Minute},                        // 40 min
				},
			},
			expected: 75.0,
		},
		{
			name: "Unordered reports - timestamps not in ascending order",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports: []db.UptimeReport{
					{Timestamp: now.Add(-40 * time.Minute), Duration: 40 * time.Minute}, // Out of order (should be before the one below)
					{Timestamp: now.Add(-80 * time.Minute), Duration: 40 * time.Minute},
					{Timestamp: now, Duration: 40 * time.Minute},
				},
			},
			expected:      0.0,
			wantError:     true,
			expectedError: ErrReportsNotInAscOrder,
		},
		{
			name: "No reports available",
			args: args{
				periodStart: now.Add(-120 * time.Minute),
				reports:     []db.UptimeReport{}, // Empty reports
			},
			expected:      0.0,
			wantError:     true,
			expectedError: ErrNoReportsAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateUpTimePercentage(tt.args.reports, tt.args.periodStart, now)

			// Check for expected error
			if tt.wantError {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err, "Expected specific error type")
				return
			}

			// No error expected
			require.NoError(t, err)

			assert.Equal(t, tt.expected, got)
		})
	}
}
