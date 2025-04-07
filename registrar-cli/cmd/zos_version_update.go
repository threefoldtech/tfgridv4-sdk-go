// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// zosVersionUpdateCmd represents the zos version update command
var zosVersionUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update zos version in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		mnemonic, err := cobraCmd.Flags().GetString("mnemonic")
		if err != nil {
			return err
		}

		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		version, err := cobraCmd.Flags().GetString("version")
		if err != nil {
			return err
		}

		safeToUpgrade, err := cobraCmd.Flags().GetBool("safe-to-upgrade")
		if err != nil {
			return err
		}

		err = cmd.UpdateVersion(mnemonic, network, version, safeToUpgrade)
		if err != nil {
			return err
		}

		log.Info().Msg("zos version is updated successfully")

		return nil
	},
}

func init() {
	zosVersionCmd.AddCommand(zosVersionUpdateCmd)
	zosVersionUpdateCmd.Flags().StringP("mnemonic", "m", "", "account mnemonic")
	zosVersionUpdateCmd.Flags().StringP("version", "v", "v0.0.0", "new zos version")
	zosVersionUpdateCmd.Flags().BoolP("safe-to-upgrade", "u", false, "safe to upgrade")
	zosVersionUpdateCmd.MarkFlagsRequiredTogether("mnemonic", "version", "safe-to-upgrade")
}
