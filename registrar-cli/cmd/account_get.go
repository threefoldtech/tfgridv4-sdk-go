// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// accountGetCmd represents the account get command
var accountGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get account from node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		twinID, err := cobraCmd.Flags().GetUint64("twin-id")
		if err != nil {
			return err
		}

		pk, err := cobraCmd.Flags().GetString("public-key")
		if err != nil {
			return err
		}

		account, err := cmd.GetAccount(network, twinID, pk)
		if err != nil {
			return err
		}

		log.Info().Any("account", account).Send()

		return nil
	},
}

func init() {
	accountCmd.AddCommand(accountGetCmd)
	accountGetCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	accountGetCmd.Flags().Uint64P("twin-id", "i", 0, "twin id")
	accountGetCmd.Flags().StringP("public-key", "k", "", "account public key")
}
