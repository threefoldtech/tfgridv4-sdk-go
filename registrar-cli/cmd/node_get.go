// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// nodeGetCmd represents the node get command
var nodeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get node from node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		nodeID, err := cobraCmd.Flags().GetUint64("node-id")
		if err != nil {
			return err
		}

		twinID, err := cobraCmd.Flags().GetUint64("twin-id")
		if err != nil {
			return err
		}

		node, err := cmd.GetNode(network, nodeID, twinID)
		if err != nil {
			return err
		}

		log.Info().Any("node", node).Send()

		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeGetCmd)
	nodeGetCmd.Flags().Uint64("node-id", 0, "node id")
	nodeGetCmd.Flags().Uint64("twin-id", 0, "twin id")
}
