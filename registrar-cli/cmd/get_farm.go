// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// getFarmCmd represents the cancel command
var getFarmCmd = &cobra.Command{
	Use:   "farm",
	Short: "get farm from node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		farmID, err := cmd.Flags().GetUint64("farm-id")
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

		if farmID != 0 {
			farm, err := cli.GetFarm(farmID)
			if err != nil {
				return err
			}
			log.Info().Any("farm", farm).Send()

		} else {
			return fmt.Errorf("you need to provide either farm id to load a farm")
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getFarmCmd)
	getFarmCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	getFarmCmd.Flags().Uint64P("farm-id", "i", 0, "farm id")
}
