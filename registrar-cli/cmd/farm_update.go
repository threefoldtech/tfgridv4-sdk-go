// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// farmUpdateCmd represents the farm update command
var farmUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update farm in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		mnemonic, err := cobraCmd.Flags().GetString("menmonic")
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

		stellarAddrss, err := cobraCmd.Flags().GetString("stellar-address")
		if err != nil {
			return err
		}

		dedicated, err := cobraCmd.Flags().GetBool("dedicated")
		if err != nil {
			return err
		}

		err = cmd.UpdateFarm(farmID, mnemonic, network, farmName, stellarAddrss, dedicated)
		if err != nil {
			return err
		}

		log.Info().Msg("farm is updated successfully")

		return nil
	},
}

func init() {
	farmCmd.AddCommand(farmUpdateCmd)
	farmUpdateCmd.Flags().StringP("menmonic", "m", "", "account menmonic")
	farmUpdateCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	farmUpdateCmd.Flags().Uint64P("farm-id", "i", 0, "farm id")
	farmUpdateCmd.Flags().String("farm-name", "", "new farm name")
	farmUpdateCmd.Flags().StringP("stellar-address", "s", "", "stellar address")
	farmUpdateCmd.Flags().BoolP("dedicated", "d", false, "farm is dedicated")
}
