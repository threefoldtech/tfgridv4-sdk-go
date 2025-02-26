// Package cmd for parsing command line arguments
package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// UpdateAccountCmd represents the cancel command
var UpdateAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "update account in node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := cmd.Flags().GetString("seed")
		if err != nil {
			return err
		}

		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		relays, err := cmd.Flags().GetStringArray("relays")
		if err != nil {
			return err
		}

		rmbEncKey, err := cmd.Flags().GetString("rmb-enc-key")
		if err != nil {
			return err
		}

		u, ok := urls[network]
		if !ok {
			return fmt.Errorf("invalid network %s", network)
		}

		if len(seed) == 0 {
			return fmt.Errorf("can not initialize registrar client with no seed")
		}

		seedBytes, err := hex.DecodeString(seed)
		if err != nil {
			return err
		}

		privateKey := ed25519.NewKeyFromSeed(seedBytes)
		cli, err := client.NewRegistrarClient(u, privateKey)
		if err != nil {
			return err
		}

		var opts []client.UpdateAccountOpts
		if len(relays) > 0 {
			opts = append(opts, client.UpdateAccountWithRelays(relays))
		}

		if len(rmbEncKey) > 0 {
			opts = append(opts, client.UpdateAccountWithRMBEncKey(rmbEncKey))
		}

		err = cli.UpdateAccount(opts...)
		if err != nil {
			return err
		}
		log.Info().Msg("account is updated successfully")

		return nil
	},
}

func init() {
	updateCmd.AddCommand(UpdateAccountCmd)
	UpdateAccountCmd.Flags().StringP("seed", "s", "", "account seed key")
	UpdateAccountCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	UpdateAccountCmd.Flags().StringArrayP("relays", "r", nil, "relays urls")
	UpdateAccountCmd.Flags().StringP("rmb-enc-key", "k", "", "rmb encryption key")
}
