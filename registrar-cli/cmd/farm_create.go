// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// farmCreateCmd represents the farm create command
var farmCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create new farm in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		mnemonic, err := cobraCmd.Flags().GetString("mnemonic")
		if err != nil {
			return errors.Wrap(err, "failed to get mnemonic flag")
		}

		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return errors.Wrap(err, "failed to get network flag")
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

		// Validate required inputs
		if farmName == "" {
			return errors.New("farm name is required (use --farm-name flag)")
		}

		if network == "" {
			return errors.New("network is required (use --network flag)")
		}

		if mnemonic == "" {
			return errors.New("mnemonic is required (use --mnemonic flag)")
		}

		farmID, err := cmd.CreateFarm(mnemonic, network, farmName, stellarAddress, dedicated)
		if err != nil {
			return errors.Wrap(err, "failed to create farm")
		}

		log.Info().Uint64("farmID", farmID).Msg("farm is created successfully")
		return nil
	},
}

func init() {
	farmCmd.AddCommand(farmCreateCmd)
	farmCreateCmd.Flags().StringP("mnemonic", "m", "", "account mnemonic")
	farmCreateCmd.Flags().StringP("farm-name", "f", "", "farm name")
	farmCreateCmd.Flags().StringP("stellar-address", "s", "", "stellar address")
	farmCreateCmd.MarkFlagsRequiredTogether("mnemonic", "farm-name", "stellar-address")

	farmCreateCmd.Flags().BoolP("dedicated", "d", false, "is farm dedicated")
}
