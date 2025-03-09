// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// getNodeCmd represents the cancel command
var getNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "get node from node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		nodeID, err := cmd.Flags().GetUint64("node-id")
		if err != nil {
			return err
		}

		twinID, err := cmd.Flags().GetUint64("twin-id")
		if err != nil {
			return err
		}

		u, ok := urls[network]
		if !ok {
			return fmt.Errorf("invalid network %s", network)
		}

		seed, err := generateRandomSeed()
		if err != nil {
			return err
		}

		seedBytes, err := hex.DecodeString(seed)
		if err != nil {
			return err
		}

		cli, err := client.NewRegistrarClient(u, seedBytes)
		if err != nil {
			return err
		}

		if nodeID != 0 {
			node, err := cli.GetNode(nodeID)
			if err != nil {
				return err
			}
			log.Info().Any("node", node).Send()

		} else if twinID != 0 {
			node, err := cli.GetNodeByTwinID(twinID)
			if err != nil {
				return err
			}
			log.Info().Any("node", node).Send()
		} else {
			return fmt.Errorf("you need to provide either twin id or node id to load a node")
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getNodeCmd)
	getNodeCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	getNodeCmd.Flags().Uint64("node-id", 0, "node id")
	getNodeCmd.Flags().Uint64("twin-id", 0, "twin id")
}
