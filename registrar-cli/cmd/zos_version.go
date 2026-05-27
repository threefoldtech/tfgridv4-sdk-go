// Package cmd for parsing command line arguments
package cmd

import (
	"github.com/spf13/cobra"
)

// zosVersionCmd represents the zos version command
var zosVersionCmd = &cobra.Command{
	Use:   "zos-version",
	Short: "zos-version command represents events on zos-version on Threefold grid4",
}

func init() {
	rootCmd.AddCommand(zosVersionCmd)
}
