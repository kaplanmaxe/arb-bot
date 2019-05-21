package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the installed helgart version",
	Long:  `broker fetches cryptocurrency markets and potentially exposes a websocket API`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Printf("helgart broker version: %s\n", version)
	},
}
