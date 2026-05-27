// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// farmGetCmd represents the farm get command
var farmGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get farm from node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return errors.Wrap(err, "failed to get network flag")
		}

		farmID, err := cobraCmd.Flags().GetUint64("farm-id")
		if err != nil {
			return errors.Wrap(err, "failed to get farm-id flag")
		}

		if farmID == 0 {
			return errors.New("farm id is required (use --farm-id flag with a non-zero value)")
		}

		farm, err := cmd.GetFarm(network, farmID)
		if err != nil {
			return errors.Wrap(err, "failed to get farm")
		}

		log.Info().Any("farm", farm).Send()
		return nil
	},
}

func init() {
	farmCmd.AddCommand(farmGetCmd)
	farmGetCmd.Flags().Uint64P("farm-id", "i", 0, "farm id")
	if err := farmGetCmd.MarkFlagRequired("farm-id"); err != nil {
		log.Fatal().Err(err).Send()
	}
}
