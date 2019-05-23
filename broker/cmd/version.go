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

func init() {
	// foo := &wsapi.ArbMarket{
	// 	HePair: "MOCK-USD",
	// 	Spread: 0.23,
	// 	Low: &wsapi.ArbMarket_ActiveMarket{
	// 		Exchange: "Kraken",
	// 		HePair:   "MOCK-USD",
	// 		ExPair:   "MOCK-USD",
	// 		Price:    "123",
	// 	},
	// 	High: &wsapi.ArbMarket_ActiveMarket{
	// 		Exchange: "Binance",
	// 		HePair:   "MOCK-USD",
	// 		ExPair:   "MOCK-USD",
	// 		Price:    "456",
	// 	},
	// }
	// b, err := proto.Marshal(foo)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// newFoo := &wsapi.ArbMarket{}
	// err = proto.Unmarshal(b, newFoo)
	// log.Fatal(newFoo)
	rootCmd.AddCommand(versionCmd)
}
