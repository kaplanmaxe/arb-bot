package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kaplanmaxe/cw-websocket/binance"
	"github.com/kaplanmaxe/cw-websocket/broker"
	"github.com/kaplanmaxe/cw-websocket/coinbase"
	"github.com/kaplanmaxe/cw-websocket/kraken"
)

func main() {
	log.Print("Starting quote server")
	interrupt := make(chan os.Signal, 1)
	quoteCh := make(chan broker.Quote)
	doneCh := make(chan struct{}, 1)
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
	kraken.Connect(ctx, quoteCh)
	coinbase.Connect(ctx, quoteCh)
	binance.Connect(ctx, quoteCh)
	go func() {
		for {
			select {
			case quote := <-quoteCh:
				log.Printf("Quote: %#v", quote)
			case <-interrupt:
				log.Println("interrupt")
				cancel()
				select {
				case <-time.After(4 * time.Second):
					close(doneCh)
					return
				}
			default:
			}
		}
	}()
	<-doneCh
}
