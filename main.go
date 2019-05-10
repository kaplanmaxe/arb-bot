package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/kaplanmaxe/cw-websocket/binance"
	"github.com/kaplanmaxe/cw-websocket/coinbase"
	"github.com/kaplanmaxe/cw-websocket/kraken"
)

func main() {
	log.Print("Starting quote server")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	// todo: add errors ch
	kraken := kraken.NewClient([]kraken.Subscription{
		{
			Type: kraken.TICKER,
			// TODO: why won't this work for only one pair?
			Pair: []string{"XBT/USD", "ETH/USD"},
		},
	})
	coinbase := coinbase.NewClient([]string{"BTC-USD", "ETH-USD"})
	binance := binance.NewClient([]string{"BTCUSD", "ETHUSD"})
	go kraken.Connect(ctx)
	go coinbase.Connect(ctx)
	go binance.Connect(ctx)
	for {
		select {
		case <-interrupt:
			cancel()
			return
		default:
		}
	}

}
