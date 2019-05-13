package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kaplanmaxe/helgart/binance"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/coinbase"
	"github.com/kaplanmaxe/helgart/exchange"
	"github.com/kaplanmaxe/helgart/kraken"
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
	// coinbase := coinbase.NewClient([]string{"BTC-USD", "ETH-USD"})
	// coinbase := coinbase.NewClient(exchange.NewSource(exchange.COINBASE, quoteCh), []string{"BTC-USD", "ETH-USD"}, quoteCh)
	coinbase := exchange.NewSource(coinbase.NewClient([]string{"BTC-USD", "ETH-USD"}, quoteCh), exchange.COINBASE, quoteCh)
	binance := binance.NewClient([]string{"BTCUSD", "ETHUSD"})
	kraken.Connect(ctx, quoteCh)
	// coinbase.Connect(ctx, quoteCh)
	coinbase.Start(ctx)
	binance.Connect(ctx, quoteCh)

	// cb := exchange.NewSource(exchange.COINBASE, quoteCh)
	// cb.Start()
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
