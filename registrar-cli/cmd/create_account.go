// Package cmd for parsing command line arguments
package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// createAccountCmd represents the cancel command
var createAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "create new account in node registrar",
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
			seed, err = generateRandomSeed()
			if err != nil {
				return err
			}
			log.Info().Msgf("New Seed (Hex): %s", seed)
		}

		seedBytes, err := hex.DecodeString(seed)
		if err != nil {
			return err
		}

		privateKey := ed25519.NewKeyFromSeed(seedBytes)
		publicKey := privateKey.Public().(ed25519.PublicKey)

		log.Info().Msgf("public key (Hex): %s", hex.EncodeToString(publicKey))

		cli, err := client.NewRegistrarClient(u, privateKey)
		if err != nil {
			return err
		}

		account, err := cli.CreateAccount(relays, rmbEncKey)
		if err != nil {
			return err
		}

		log.Info().Uint64("twinID", account.TwinID).Msg("account is created successfully")

		return nil
	},
}

func init() {
	createCmd.AddCommand(createAccountCmd)
	createAccountCmd.Flags().StringP("seed", "s", "", "account seed key")
	createAccountCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	createAccountCmd.Flags().StringArrayP("relays", "r", nil, "relays urls")
	createAccountCmd.Flags().StringP("rmb-enc-key", "k", "", "rmb encryption key")
}

func generateRandomSeed() (string, error) {
	s := make([]byte, 32)
	_, err := rand.Read(s)
	if err != nil {
		return "", err
	}

	seed := hex.EncodeToString(s)
	return seed, nil
}
