// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "account command represents events on account on Threefold grid4",
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
