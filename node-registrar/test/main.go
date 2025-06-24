package main

import (
	"fmt"
	"log"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

func main() {
	// Create a new client with the registrar server URL
	registrarClient, err := client.NewRegistrarClient("http://localhost:8080/api/v1")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get node capacity rewards by node ID
	nodeID := uint64(34) // Replace with actual node ID
	rewards, err := registrarClient.GetNodeCapacityRewards(nodeID)
	if err != nil {
		log.Fatalf("Failed to get node capacity rewards: %v", err)
	}

	// Print the rewards information
	fmt.Printf("Node %d Rewards:\n", nodeID)
	fmt.Printf("Total: %f\n", rewards.Total)
	fmt.Printf("Farmer Reward: %f\n", rewards.FarmerReward)
	fmt.Printf("TF Reward: %f\n", rewards.TfReward)
	fmt.Printf("FP Reward: %f\n", rewards.FpReward)
	fmt.Printf("Uptime Percentage: %f%%\n", rewards.UpTimePercentage)
}
