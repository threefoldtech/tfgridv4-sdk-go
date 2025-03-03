// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// farmCreateCmd represents the farm create command
var farmCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create new farm in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		seed, err := cobraCmd.Flags().GetString("seed")
		if err != nil {
			return err
		}

		network, err := cobraCmd.Flags().GetString("network")
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

		farmID, err := cmd.CreaeteFarm(seed, network, farmName, dedicated)
		if err != nil {
			return err
		}

		log.Info().Uint64("farmID", farmID).Msg("farm is created successfully")

		return nil
	},
}

func init() {
	farmCmd.AddCommand(farmCreateCmd)
	farmCreateCmd.Flags().StringP("seed", "s", "", "account seed key")
	farmCreateCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	farmCreateCmd.Flags().StringP("farm-name", "f", "", "farm name")
	farmCreateCmd.Flags().BoolP("dedicated", "d", false, "is farm dedicated")
}
