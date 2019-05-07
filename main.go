package main

import (
	"log"

	"github.com/kaplanmaxe/cw-websocket/kraken"
)

func main() {
	log.Print("Starting quote server")
	subs := []kraken.Subscription{
		{
			Type: kraken.TICKER,
			// TODO: why won't this work for only one pair?
			Pair: []string{"XBT/USD", "ETH/USD"},
		},
	}
	client := kraken.NewClient(subs)
	client.Connect()
}
