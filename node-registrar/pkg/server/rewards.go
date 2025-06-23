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
var ErrInvalidUptimePercentage = errors.New("invalid uptime percentage")

type Reward struct {
	FarmerReward     float64 //FarmerReward: the reward for the node owner
	TfReward         float64 //TfReward: the reward for the Threefold Foundation
	FpReward         float64 //FpReward: the reward for the Farming Pool
	Total            float64 //Total: the total reward
	UpTimePercentage float64 //UpTimePercentage: the uptime percentage of the node
}

// CalculateMonthlyReward calculates the monthly reward in INCA for a given node capacity.
//
// The rewards are calculated as follows:
//
// - Certified capacity rewards factor:
//   - Memory: 8.0 INCA per GB
//   - SSD: 31.5 INCA per TB
//   - HDD: 7.0 INCA per TB
//
// - Reward distribution:
//   - Farmer: 60% of the reward
//   - Threefold Foundation: 20% of the reward
//   - Farming Pool: 20% of the reward
//
// CalculateMonthlyReward returns @Reward struct
//
// CalculateMonthlyReward takes the following parameters:
//
// - capacity: the node capacity, in form of a db.Resources Type
// - upTimePercentage: the uptime percentage of the node, as a float64 value between 0 and 100
//
// - Note: if the uptime percentage is less than 90, the node will not reserve any rewards.

func CalculateMonthlyReward(capacity db.Resources, upTimePercentage float64) (Reward, error) {
	if upTimePercentage < 0 || upTimePercentage > 100 {
		return Reward{}, ErrInvalidUptimePercentage
	}
	if upTimePercentage < 90 {
		return Reward{UpTimePercentage: upTimePercentage}, nil
	}

	total := (bytesToGB(capacity.MRU)*MemoryRewardPerGB + bytesToTB(capacity.SRU)*SsdRewardPerTB + bytesToTB(capacity.HRU)*HddRewardPerTB) * (upTimePercentage / 100)

	return Reward{
		FarmerReward:     truncateFloat(total*FarmerRewardPercentage, 3),
		TfReward:         truncateFloat(total*TfRewardPercentage, 3),
		FpReward:         truncateFloat(total*FpRewardPercentage, 3),
		Total:            truncateFloat(total, 3),
		UpTimePercentage: upTimePercentage,
	}, nil
}

func bytesToGB(bytes uint64) float64 {
	return float64(bytes) / math.Pow(1024, 3)
}

func bytesToTB(bytes uint64) float64 {
	return float64(bytes) / math.Pow(1024, 4)
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
//
// Note: This function assumes that the reports are ordered by timestamp in ascending order.
//
// Parameters:
//   - reports: a slice of UptimeReport
//   - periodStart: the start of the period
//   - now: the current time
//
// Returns:
//   - a float64 representing the uptime percentage
func calculateUpTimePercentage(reports []db.UptimeReport, periodStart, now time.Time) (float64, error) {

	if len(reports) == 0 {
		return 0.0, nil
	}

	for i := 0; i < len(reports)-1; i++ {
		if reports[i].Timestamp.After(reports[i+1].Timestamp) {
			return 0.0, errors.New("timestamps are not ordered correctly")
		}
	}

	//append starter point
	reports = append([]db.UptimeReport{
		{
			Timestamp: periodStart,
			Duration:  time.Duration(0),
		},
	}, reports...)

	var downtime time.Duration = 0
	for i := 0; i < len(reports)-1; i++ {

		curr := reports[i]
		next := reports[i+1]

		curr.Duration = time.Duration(curr.Duration * time.Second)
		next.Duration = time.Duration(next.Duration * time.Second)

		expected := next.Timestamp.Sub(curr.Timestamp).Truncate(time.Second)
		actual := next.Duration.Truncate(time.Second)
		if curr.Duration > next.Duration || actual < expected {
			downtime += expected - actual
		}
	}
	// if there is a gap equal
	// s or larger than th @UPTIME_EVENTS_INTERVAL between the last report and now, add it to the downtime
	elapsedSinceLast := now.Sub(reports[len(reports)-1].Timestamp).Truncate(time.Second)
	if elapsedSinceLast.Seconds() >= UptimeEventsInterval {
		downtime += elapsedSinceLast
	}
	return truncateFloat(float64(now.Sub(periodStart)-downtime)/float64(now.Sub(periodStart))*100, 2), nil
}

// calculatePeriodStart returns the start of the period that contains the reference time.
//
// The function uses the unix timestamp of the first period start (FirstPeriodStartTimestamp) and the standard period duration (StandardPeriodDuration) to calculate the start of the period.
//
// Parameter:
//   - referenceTime: the reference time used to calculate its period start time
func calculatePeriodStart(referenceTime time.Time) time.Time {
	secondsSinceFirstPeriod := referenceTime.Unix() - FirstPeriodStartTimestamp
	periodOffset := secondsSinceFirstPeriod % StandardPeriodDuration
	periodStart := referenceTime.Unix() - periodOffset
	return time.Unix(periodStart, 0)
}

func truncateFloat(num float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Trunc(num*pow) / pow
}
