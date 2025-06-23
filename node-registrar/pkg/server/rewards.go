package server

import (
	"errors"
	"math"
	"time"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

const (
	MemoryRewardPerGB float64 = 8.0  // Certified capacity rewards factor Per GB
	SsdRewardPerTB    float64 = 31.5 // Certified capacity rewards factor Per TB
	HddRewardPerTB    float64 = 7.0  // Certified capacity rewards factor Per TB

	FarmerRewardPercentage float64 = 0.6 // FarmerRewardPercentage is the percentage of the reward that goes to the node owner (60%)
	TfRewardPercentage     float64 = 0.2 // TfRewardPercentage is the percentage of the reward that goes to the Threefold  (20%)
	FpRewardPercentage     float64 = 0.2 // FpRewardPercentage is the percentage of the reward that goes to the Farming Pool (20%)

	// defines the number of decimal places to keep in reward calculations
	RewardPrecisionDecimalPlaces int = 3

	// MinUptimePercentageForReward is the minimum uptime percentage required for a node to receive rewards
	MinUptimePercentageForReward float64 = 90.0
)

// Time related constants
const (
	// TODO update the start timestamp
	FirstPeriodStartTimestamp int64 = 1522501000

	// Uptime events are supposed to happen every 40 minutes.
	// Here we set this to one hour (3600 sec) to allow some room.
	UptimeEventsInterval = 3600

	// The duration of a standard period, as used by the minting payouts, in seconds.
	// Calculated as: 24 hours * 60 minutes * 60 seconds * (365*3 + 366*2) / 60 periods
	StandardPeriodDuration int64 = 24 * 60 * 60 * (365*3 + 366*2) / 60
)

// Error messages
var (
	ErrInvalidUptimePercentage = errors.New("invalid uptime percentage")
	ErrReportsNotInAscOrder    = errors.New("timestamps are not ordered correctly")
	ErrNoReportsAvailable      = errors.New("no reports available For this node")
)

type Reward struct {
	FarmerReward     float64 //FarmerReward: the reward for the node owner
	TfReward         float64 //TfReward: the reward for the Threefold Foundation
	FpReward         float64 //FpReward: the reward for the Farming Pool
	Total            float64 //Total: the total reward
	UpTimePercentage float64 //UpTimePercentage: the uptime percentage of the node
}

// CalculateCapacityReward calculates the reward in INCA for a given node capacity.
//
// - Note: if the uptime percentage is less than MinUptimePercentageForReward, the node will not receive any rewards.
func CalculateCapacityReward(capacity db.Resources, upTimePercentage float64) (Reward, error) {
	if upTimePercentage < 0 || upTimePercentage > 100 {
		return Reward{}, ErrInvalidUptimePercentage
	}
	if upTimePercentage < MinUptimePercentageForReward {
		return Reward{UpTimePercentage: upTimePercentage}, nil
	}

	total := (bytesToGB(capacity.MRU)*MemoryRewardPerGB + bytesToTB(capacity.SRU)*SsdRewardPerTB + bytesToTB(capacity.HRU)*HddRewardPerTB) * (upTimePercentage / 100)

	return Reward{
		FarmerReward:     truncateFloat(total*FarmerRewardPercentage, RewardPrecisionDecimalPlaces),
		TfReward:         truncateFloat(total*TfRewardPercentage, RewardPrecisionDecimalPlaces),
		FpReward:         truncateFloat(total*FpRewardPercentage, RewardPrecisionDecimalPlaces),
		Total:            truncateFloat(total, RewardPrecisionDecimalPlaces),
		UpTimePercentage: upTimePercentage,
	}, nil
}

// bytesToGB converts bytes to gigabytes.
func bytesToGB(bytes uint64) float64 {
	return float64(bytes) / math.Pow(1024, 3)
}

// bytesToTB converts bytes to terabytes.
func bytesToTB(bytes uint64) float64 {
	return float64(bytes) / math.Pow(1024, 4)
}

// calculatePeriodStart returns the start of the period that contains the reference time.
//
// The function uses the unix timestamp of the first period start (FirstPeriodStartTimestamp) and the standard period duration (StandardPeriodDuration) to calculate the start of the period.
func calculatePeriodStart(referenceTime time.Time) time.Time {
	secondsSinceFirstPeriod := referenceTime.Unix() - FirstPeriodStartTimestamp
	periodOffset := secondsSinceFirstPeriod % StandardPeriodDuration
	periodStart := referenceTime.Unix() - periodOffset
	return time.Unix(periodStart, 0)
}

// truncateFloat truncates a floating point number to the specified precision.
func truncateFloat(num float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Trunc(num*pow) / pow
}

// calculateDowntimeFromReports calculates the downtime from a sequence of uptime reports.
//
// This function iterates through the reports and calculates downtime by comparing the
// gap between consecutive timestamps with the reported duration.
// If the duration is less than the gap or if the current duration is greater than the next duration,
// the difference is counted as downtime.
func calculateDowntimeFromReports(reports []db.UptimeReport) time.Duration {
	var downtime time.Duration
	for i := range len(reports) - 1 {
		curr := reports[i]
		next := reports[i+1]
		gapBetweenTimeStamps := next.Timestamp.Sub(curr.Timestamp)
		duration := next.Duration
		if curr.Duration > next.Duration || duration < gapBetweenTimeStamps {
			downtime += gapBetweenTimeStamps - duration
		}
	}
	return downtime
}

// areReportsOrderedCorrectly verifies that uptime reports are ordered chronologically by timestamp.
func areReportsOrderedCorrectly(reports []db.UptimeReport) bool {
	for i := range len(reports) - 1 {
		if reports[i].Timestamp.After(reports[i+1].Timestamp) {
			return false
		}
	}
	return true
}

// calculatePercentage calculates the percentage of uptime based on total period duration and downtime.
func calculatePercentage(totalPeriod, downtime time.Duration) float64 {
	// Calculate actual uptime by subtracting downtime from total period
	actualUptime := totalPeriod - downtime

	// Calculate the percentage (actual uptime / total period * 100)
	uptimeRatio := float64(actualUptime) / float64(totalPeriod)
	uptimePercentage := uptimeRatio * 100

	// Truncate to 2 decimal places
	return truncateFloat(uptimePercentage, 2)
}

// downtimeSinceLastReportTimestamp calculates the downtime since the last uptime report timestamp.
//
// This function determines if there has been a significant gap (exceeding UptimeEventsInterval)
// since the last report. If such a gap exists, it's considered downtime.
func downtimeSinceLastReportTimestamp(lastReportTimestamp time.Time, currentTime time.Time) time.Duration {
	elapsedSinceLast := currentTime.Sub(lastReportTimestamp).Truncate(time.Second)

	if elapsedSinceLast.Seconds() >= UptimeEventsInterval {
		return elapsedSinceLast
	}
	return 0
}

// calculateUpTimePercentage calculates the uptime percentage for a given node within a specific period.
//
// This function takes a slice of UptimeReport, a period start time and a current time as parameters.
// It calculates the uptime percentage by comparing the expected uptime (calculated by subtracting the timestamp of the previous report from the current report)
// with the actual uptime (calculated from the duration the current report).
// If the actual uptime is less than the expected uptime, the difference is counted as downtime.
// Additionally, if there is a gap equals or larger than the @UPTIME_EVENTS_INTERVAL between the last report and now, add it to the downtime.
// The uptime percentage is then calculated by subtracting the total downtime from the total elapsed time since the period start and dividing the result by the total elapsed time.
// The result is then multiplied by 100 to get the percentage.
func calculateUpTimePercentage(reports []db.UptimeReport, periodStart, now time.Time) (float64, error) {

	if len(reports) == 0 {
		return 0.0, ErrNoReportsAvailable
	}

	if !areReportsOrderedCorrectly(reports) {
		return 0.0, ErrReportsNotInAscOrder
	}

	//append starter point
	reports = append([]db.UptimeReport{
		{
			Timestamp: periodStart,
			Duration:  time.Duration(0),
		},
	}, reports...)

	var downtime time.Duration = 0
	downtime += calculateDowntimeFromReports(reports)
	downtime += downtimeSinceLastReportTimestamp(reports[len(reports)-1].Timestamp, now)

	totalPeriod := now.Sub(periodStart)
	return calculatePercentage(totalPeriod, downtime), nil
}
