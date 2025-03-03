// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// getZosVersionCmd represents the zos version command
var getZosVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "get zos version from Threefold grid4",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		version, err := cmd.GetVersion(network)
		if err != nil {
			return err
		}

		log.Info().Any("zosVersion", version).Send()

		return nil
	},
}

func init() {
	getCmd.AddCommand(getZosVersionCmd)
	getZosVersionCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
}
