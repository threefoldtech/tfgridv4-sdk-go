// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// createFarmCmd represents the cancel command
var createFarmCmd = &cobra.Command{
	Use:   "farm",
	Short: "create new farm in node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := cmd.Flags().GetString("seed")
		if err != nil {
			return err
		}

		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		farmName, err := cmd.Flags().GetString("farm-name")
		if err != nil {
			return err
		}

		dedicated, err := cmd.Flags().GetBool("dedicated")
		if err != nil {
			return err
		}

		u, ok := urls[network]
		if !ok {
			return fmt.Errorf("invalid network %s", network)
		}

		if len(seed) == 0 {
			return fmt.Errorf("can not initialize registrar client with no seed")
		}

		seedBytes, err := hex.DecodeString(seed)
		if err != nil {
			return err
		}

		cli, err := client.NewRegistrarClient(u, seedBytes)
		if err != nil {
			return err
		}

		farm, err := cli.CreateFarm(farmName, dedicated)
		if err != nil {
			return err
		}

		log.Info().Uint64("farmID", farm).Msg("farm is created successfully")

		return nil
	},
}

func init() {
	createCmd.AddCommand(createFarmCmd)
	createFarmCmd.Flags().StringP("seed", "s", "", "account seed key")
	createFarmCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	createFarmCmd.Flags().StringP("farm-name", "f", "", "farm name")
	createFarmCmd.Flags().BoolP("dedicated", "d", false, "is farm dedicated")
}
