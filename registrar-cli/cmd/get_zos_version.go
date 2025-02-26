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

// getZosVersionCmd represents the zos version command
var getZosVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "get zos version from Threefold grid4",
	RunE: func(cmd *cobra.Command, args []string) error {
		network, err := cmd.Flags().GetString("network")
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

		privateKey := ed25519.NewKeyFromSeed(seedBytes)
		cli, err := client.NewRegistrarClient(u, privateKey)
		if err != nil {
			return err
		}

		version, err := cli.GetZosVersion()
		if err != nil {
			return err
		}
		log.Info().Any("zosVersion", version).Send()

		return nil
	},
}

func init() {
	getCmd.AddCommand(getZosVersionCmd)
	getZosVersionCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
}
