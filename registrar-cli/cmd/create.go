// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create account/farm on Threefold grid4",
}

func init() {
	rootCmd.AddCommand(createCmd)
}
