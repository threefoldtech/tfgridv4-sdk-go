// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/cmd"
)

// accountUpdateCmd represents the account update command
var accountUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update account in node registrar",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		seed, err := cobraCmd.Flags().GetString("menmonic")
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

		err = cmd.UpdateAccount(seed, network, relays, rmbEncKey)
		if err != nil {
			return err
		}

		log.Info().Msg("account is updated successfully")

		return nil
	},
}

func init() {
	accountCmd.AddCommand(accountUpdateCmd)
	accountUpdateCmd.Flags().StringP("menmonic", "m", "", "account menmonic")
	accountUpdateCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	accountUpdateCmd.Flags().StringArrayP("relays", "r", nil, "relays urls")
	accountUpdateCmd.Flags().StringP("rmb-enc-key", "k", "", "rmb encryption key")
}
