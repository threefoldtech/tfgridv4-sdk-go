// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// createCmd represents the farm command
var farmCmd = &cobra.Command{
	Use:   "farm",
	Short: "farm command represents events on farm on Threefold grid4",
}

func init() {
	rootCmd.AddCommand(farmCmd)
}
