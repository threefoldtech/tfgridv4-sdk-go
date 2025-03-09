// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get account/farm/node from Threefold grid4",
}

func init() {
	rootCmd.AddCommand(getCmd)
}
