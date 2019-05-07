package main

import (
	"fmt"

	"github.com/kaplanmaxe/cw-websocket/kraken"
	"github.com/kaplanmaxe/cw-websocket/kraken/transport"
)

func main() {
	fmt.Println("Hello world")
	subs := []kraken.Subscription{
		{
			Type: kraken.TICKER,
			Pair: "XBT/USD",
		},
		{
			Type: kraken.TICKER,
			Pair: "XBT/EUR",
		},
	}
	client := transport.NewClient(subs)
	client.Connect()
}
