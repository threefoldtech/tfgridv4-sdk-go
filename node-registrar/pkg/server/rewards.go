package server

import (
	"errors"
	"math"
	"time"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

const (

	// Certified capacity rewards factor
	MEMORY_REWARD_PER_GB = 8.0  // INCA per GB
	SSD_REWARD_PER_TB    = 31.5 // INCA per TB
	HDD_REWARD_PER_TB    = 7.0  // INCA per TB

	// Reward distribution
	FARMER_REWARD_PERCENTAGE = 0.6 // 60% of the reward goes to the node owner
	TF_REWARD_PERCENTAGE     = 0.2 // 20% of the reward goes to the Threefold Foundation
	FP_REWARD_PERCENTAGE     = 0.2 // 20% of the reward goes to the Farming Pool

	// TODO update the start timestamp
	FIRST_PERIOD_START_TIMESTAMP int64 = 1522501000

	// uptime events are supposed to happen every 40 minutes.
	// here we set this to one hour (3600 sec) to allow some room.
	UPTIME_EVENTS_INTERVAL = 3600

	// The duration of a standard period, as used by the minting payouts, in seconds.
	STANDARD_PERIOD_DURATION int64 = 24 * 60 * 60 * (365*3 + 366*2) / 60
)

// Error messages
var ErrInvalidUptimePercentage = errors.New("invalid uptime percentage")

type Reward struct {
	FarmerReward     float64
	TFReward         float64
	FPReward         float64
	Total            float64
	UpTimePercentage float64
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
// CalculateMonthlyReward returns the following values:
//   - FarmerReward: the reward for the node owner
//   - TFReward: the reward for the Threefold Foundation
//   - FPReward: the reward for the Farming Pool
//   - Total: the total reward
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
		return Reward{
			FarmerReward: 0,
			TFReward:     0,
			FPReward:     0,
			Total:        0,
		}, nil
	}

	total := (bytesToGB(capacity.MRU)*MEMORY_REWARD_PER_GB + bytesToTB(capacity.SRU)*SSD_REWARD_PER_TB + bytesToTB(capacity.HRU)*HDD_REWARD_PER_TB) * (upTimePercentage / 100)

	return Reward{
		FarmerReward:     total * FARMER_REWARD_PERCENTAGE,
		TFReward:         total * TF_REWARD_PERCENTAGE,
		FPReward:         total * FP_REWARD_PERCENTAGE,
		Total:            total,
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
func calculateUpTimePercentage(reports []db.UptimeReport, periodStart, now time.Time) float64 {

	if len(reports) == 0 {
		return 0.0
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

		//TODO should we check the order of timestamp?

		expected := next.Timestamp.Sub(curr.Timestamp).Truncate(time.Second)
		actual := next.Duration.Truncate(time.Second)
		if curr.Duration > next.Duration || actual < expected {
			downtime += expected - actual
		}
	}
	// if there is a gap equals or larger than th @UPTIME_EVENTS_INTERVAL between the last report and now, add it to the downtime
	elapsedSinceLast := now.Sub(reports[len(reports)-1].Timestamp).Truncate(time.Second)
	if elapsedSinceLast.Seconds() >= UPTIME_EVENTS_INTERVAL {
		downtime += elapsedSinceLast
	}
	return truncateFloat(float64(now.Sub(periodStart)-downtime)/float64(now.Sub(periodStart))*100, 2)
}

// calculateCurrentPeriodStart returns the start of the current period.
//
// The function uses the unix timestamp of the first period start (FIRST_PERIOD_START_TIMESTAMP) and the standard period duration (STANDARD_PERIOD_DURATION) to calculate the start of the current period.
//
// Parameter:
//   - now: the reference time used to calculate the current period start
func calculateCurrentPeriodStart(now time.Time) time.Time {
	secondsSinceFirstPeriod := now.Unix() - FIRST_PERIOD_START_TIMESTAMP
	periodOffset := secondsSinceFirstPeriod % STANDARD_PERIOD_DURATION
	currentPeriodStart := now.Unix() - periodOffset
	return time.Unix(currentPeriodStart, 0)
}

func truncateFloat(num float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Trunc(num*pow) / pow
}
