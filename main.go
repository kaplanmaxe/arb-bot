package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kaplanmaxe/helgart/binance"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/coinbase"
	"github.com/kaplanmaxe/helgart/exchange"
)

func main() {
	log.Print("Starting quote server")
	interrupt := make(chan os.Signal, 1)
	quoteCh := make(chan broker.Quote)
	doneCh := make(chan struct{}, 1)
	signal.Notify(interrupt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	// todo: add errors ch
	// kraken := kraken.NewClient([]kraken.Subscription{
	// 	{
	// 		Type: kraken.TICKER,
	// 		// TODO: why won't this work for only one pair?
	// 		Pair: []string{"XBT/USD", "ETH/USD"},
	// 	},
	// })
	coinbase := coinbase.NewClient([]string{"BTC-USD", "ETH-USD"}, api.NewSource(exchange.COINBASE), quoteCh)
	binance := binance.NewClient([]string{}, api.NewSource(exchange.BINANCE), quoteCh)
	// kraken.Connect(ctx, quoteCh)
	coinbase.Start(ctx)
	binance.Start(ctx)
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
