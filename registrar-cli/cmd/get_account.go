// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// getAccountCmd represents the cancel command
var getAccountCmd = &cobra.Command{
	Use:   "account",
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
	getCmd.AddCommand(getAccountCmd)
	getAccountCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	getAccountCmd.Flags().Uint64P("twin-id", "i", 0, "twin id")
	getAccountCmd.Flags().StringP("public-key", "k", "", "account public key")
}
