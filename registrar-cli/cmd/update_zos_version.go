// Package cmd for parsing command line arguments
package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

// UpdateZosVersionCmd represents the update farm command
var UpdateZosVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "update zos version in node registrar",
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := cmd.Flags().GetString("seed")
		if err != nil {
			return err
		}

		network, err := cmd.Flags().GetString("network")
		if err != nil {
			return err
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			return err
		}

		safeToUpgrade, err := cmd.Flags().GetBool("safe-to-upgrade")
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

		cli, err := client.NewRegistrarClient(u, seedBytes)
		if err != nil {
			return err
		}

		err = cli.SetZosVersion(version, safeToUpgrade)
		if err != nil {
			return err
		}
		log.Info().Msg("farm is updated successfully")

		return nil
	},
}

func init() {
	updateCmd.AddCommand(UpdateZosVersionCmd)
	UpdateZosVersionCmd.Flags().StringP("seed", "s", "", "account seed key")
	UpdateZosVersionCmd.Flags().StringP("network", "n", "", "network (dev, qa, test, main)")
	UpdateZosVersionCmd.Flags().StringP("version", "v", "v0.0.0", "new zos version")
	UpdateZosVersionCmd.Flags().BoolP("safe-to-upgrade", "u", false, "safe to upgrade")
}
