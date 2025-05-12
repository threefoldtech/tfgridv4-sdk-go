// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// farmUpdateCmd represents the farm update command
var farmUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update farm in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		mnemonic, err := cobraCmd.Flags().GetString("mnemonic")
		if err != nil {
			return errors.Wrap(err, "failed to get mnemonic flag")
		}

		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return errors.Wrap(err, "failed to get network flag")
		}

		farmID, err := cobraCmd.Flags().GetUint64("farm-id")
		if err != nil {
			return errors.Wrap(err, "failed to get farm-id flag")
		}

		farmName, err := cobraCmd.Flags().GetString("farm-name")
		if err != nil {
			return errors.Wrap(err, "failed to get farm-name flag")
		}

		stellarAddress, err := cobraCmd.Flags().GetString("stellar-address")
		if err != nil {
			return errors.Wrap(err, "failed to get stellar-address flag")
		}

		dedicated, err := cobraCmd.Flags().GetBool("dedicated")
		if err != nil {
			return errors.Wrap(err, "failed to get dedicated flag")
		}

		err = cmd.UpdateFarm(farmID, mnemonic, network, farmName, stellarAddress, dedicated)
		if err != nil {
			return errors.Wrap(err, "failed to update farm")
		}

		log.Info().Msg("farm is updated successfully")
		return nil
	},
}

func init() {
	farmCmd.AddCommand(farmUpdateCmd)
	farmUpdateCmd.Flags().StringP("mnemonic", "m", "", "account mnemonic")
	farmUpdateCmd.Flags().Uint64P("farm-id", "i", 0, "farm id")
	farmUpdateCmd.Flags().String("farm-name", "", "new farm name")
	farmUpdateCmd.MarkFlagsRequiredTogether("mnemonic", "farm-id", "farm-name")

	farmUpdateCmd.Flags().StringP("stellar-address", "s", "", "stellar address")
	farmUpdateCmd.Flags().BoolP("dedicated", "d", false, "farm is dedicated")
	farmUpdateCmd.MarkFlagsOneRequired("stellar-address", "dedicated")
}
