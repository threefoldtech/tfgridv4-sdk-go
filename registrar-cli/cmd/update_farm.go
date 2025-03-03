// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// UpdateFarmCmd represents the update farm command
var UpdateFarmCmd = &cobra.Command{
	Use:   "farm",
	Short: "update farm in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		seed, err := cobraCmd.Flags().GetString("seed")
		if err != nil {
			return err
		}

		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		farmID, err := cobraCmd.Flags().GetUint64("farm-id")
		if err != nil {
			return err
		}

		farmName, err := cobraCmd.Flags().GetString("farm-name")
		if err != nil {
			return err
		}

		dedicated, err := cobraCmd.Flags().GetBool("dedicated")
		if err != nil {
			return err
		}

		err = cmd.UpdateFarm(seed, network, farmID, farmName, dedicated)
		if err != nil {
			return err
		}

		log.Info().Msg("farm is updated successfully")

		return nil
	},
}

func init() {
	updateCmd.AddCommand(UpdateFarmCmd)
	UpdateFarmCmd.Flags().StringP("seed", "s", "", "account seed key")
	UpdateFarmCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	UpdateFarmCmd.Flags().Uint64P("farm-id", "i", 0, "farm id")
	UpdateFarmCmd.Flags().String("farm-name", "", "new farm name")
	UpdateFarmCmd.Flags().BoolP("dedicated", "d", false, "farm is dedicated")
}
