package server

import (
	"errors"
	"math"

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

)

// Error messages
var ErrInvalidUptimePercentage = errors.New("invalid uptime percentage")

type Reward struct {
	FarmerReward float64
	TFReward     float64
	FPReward     float64
	Total        float64
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
		FarmerReward: total * FARMER_REWARD_PERCENTAGE,
		TFReward:     total * TF_REWARD_PERCENTAGE,
		FPReward:     total * FP_REWARD_PERCENTAGE,
		Total:        total,
	}, nil
}

func bytesToGB(bytes uint64) float64 {
	return float64(bytes) / math.Pow(1024, 3)
}

func bytesToTB(bytes uint64) float64 {
	return float64(bytes) / math.Pow(1024, 4)
}
