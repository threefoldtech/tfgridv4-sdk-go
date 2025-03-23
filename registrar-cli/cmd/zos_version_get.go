// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// zosVersionGetCmd represents the zos version get command
var zosVersionGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get zos version from node registrar",
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
	zosVersionCmd.AddCommand(zosVersionGetCmd)
}
