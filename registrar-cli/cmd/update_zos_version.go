// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// UpdateZosVersionCmd represents the update farm command
var UpdateZosVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "update zos version in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		seed, err := cobraCmd.Flags().GetString("seed")
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

		err = cmd.UpdateVersion(seed, network, version, safeToUpgrade)
		if err != nil {
			return err
		}

		log.Info().Msg("farm is updated successfully")

		return nil
	},
}

func init() {
	updateCmd.AddCommand(UpdateZosVersionCmd)
	UpdateZosVersionCmd.Flags().StringP("seed", "s", "", "account seed key")
	UpdateZosVersionCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	UpdateZosVersionCmd.Flags().StringP("version", "v", "v0.0.0", "new zos version")
	UpdateZosVersionCmd.Flags().BoolP("safe-to-upgrade", "u", false, "safe to upgrade")
}
