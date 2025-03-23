// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// accountCreateCmd represents the account create command
var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create new account in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		mnemonic, err := cobraCmd.Flags().GetString("mnemonic")
		if err != nil {
			return err
		}

		network, err := cobraCmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		relays, err := cobraCmd.Flags().GetStringArray("relays")
		if err != nil {
			return err
		}

		rmbEncKey, err := cobraCmd.Flags().GetString("rmb-enc-key")
		if err != nil {
			return err
		}

		account, mnemonic, err := cmd.CreateAccount(network, relays, rmbEncKey, mnemonic)
		if err != nil {
			return err
		}
		log.Info().Str("mnemonic", mnemonic).Msg("new account is created with mnemonic")

		log.Info().Uint64("twinID", account.TwinID).Msg("account is created successfully")

		return nil
	},
}

func init() {
	accountCmd.AddCommand(accountCreateCmd)
	accountCreateCmd.Flags().StringP("mnemonic", "m", "", "account mnemonic")
	accountCreateCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	accountCreateCmd.Flags().StringArrayP("relays", "r", nil, "relays urls")
	accountCreateCmd.Flags().StringP("rmb-enc-key", "k", "", "rmb encryption key")
}
