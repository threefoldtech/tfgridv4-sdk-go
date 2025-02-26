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

// UpdateFarmCmd represents the update farm command
var UpdateFarmCmd = &cobra.Command{
	Use:   "farm",
	Short: "update farm in node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := cmd.Flags().GetString("seed")
		if err != nil {
			return err
		}

		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		farmID, err := cmd.Flags().GetUint64("farm-id")
		if err != nil {
			return err
		}

		farmName, err := cmd.Flags().GetString("farm-name")
		if err != nil {
			return err
		}

		dedicated, err := cmd.Flags().GetBool("dedicated")
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

		var opts []client.UpdateFarmOpts
		if len(farmName) > 0 {
			opts = append(opts, client.UpdateFarmWithName(farmName))
		}

		if dedicated {
			opts = append(opts, client.UpdateFarmWithDedicated())
		}

		err = cli.UpdateFarm(farmID, opts...)
		if err != nil {
			return err
		}
		log.Info().Msg("farm is updated successfully")

		return nil
	},
}

func init() {
	updateCmd.AddCommand(UpdateFarmCmd)
	UpdateFarmCmd.Flags().StringP("seed", "s", "", "account seed key")
	UpdateFarmCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	UpdateFarmCmd.Flags().Uint64P("farm-id", "i", 0, "farm id")
	UpdateFarmCmd.Flags().String("farm-name", "", "new farm name")
	UpdateFarmCmd.Flags().BoolP("dedicated", "d", false, "farm is dedicated")
}
