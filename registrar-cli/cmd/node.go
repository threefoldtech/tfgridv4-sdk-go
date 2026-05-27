// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "node command represents events on node on Threefold grid4",
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
