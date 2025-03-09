// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// getAccountCmd represents the cancel command
var getAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "get account from node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		twinID, err := cmd.Flags().GetUint64("twin-id")
		if err != nil {
			return err
		}

		pk, err := cmd.Flags().GetString("public-key")
		if err != nil {
			return err
		}

		u, ok := urls[network]
		if !ok {
			return fmt.Errorf("invalid network %s", network)
		}

		seed, err := generateRandomSeed()
		if err != nil {
			return err
		}

		seedBytes, err := hex.DecodeString(seed)
		if err != nil {
			return err
		}

		publicKey, err := hex.DecodeString(pk)
		if err != nil {
			return err
		}

		cli, err := client.NewRegistrarClient(u, seedBytes)
		if err != nil {
			return err
		}

		if twinID != 0 {
			account, err := cli.GetAccount(twinID)
			if err != nil {
				return err
			}
			log.Info().Any("account", account).Send()

		} else if len(publicKey) != 0 {
			account, err := cli.GetAccountByPK(publicKey)
			if err != nil {
				return err
			}
			log.Info().Any("account", account).Send()
		} else {
			return fmt.Errorf("you need to provide either twin id or public key to load an account")
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getAccountCmd)
	getAccountCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	getAccountCmd.Flags().Uint64P("twin-id", "i", 0, "twin id")
	getAccountCmd.Flags().StringP("public-key", "k", "", "account public key")
}
