package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kaplanmaxe/helgart/binance"
	"github.com/kaplanmaxe/helgart/kraken"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/coinbase"
	"github.com/kaplanmaxe/helgart/exchange"
)

func main() {
	log.Print("Starting quote server")
	interrupt := make(chan os.Signal, 1)
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error)
	doneCh := make(chan struct{}, 1)
	signal.Notify(interrupt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	kraken := kraken.NewClient([]string{"XBT/USD", "ETH/USD"}, api.NewSource(exchange.KRAKEN), quoteCh, errorCh)
	coinbase := coinbase.NewClient([]string{"BTC-USD", "ETH-USD"}, api.NewSource(exchange.COINBASE), quoteCh, errorCh)
	binance := binance.NewClient([]string{}, api.NewSource(exchange.BINANCE), quoteCh, errorCh)

	kraken.Start(ctx)
	coinbase.Start(ctx)
	binance.Start(ctx)

	go func() {
		for {
			select {
			case quote := <-quoteCh:
				log.Printf("Quote: %#v", quote)
			case err := <-errorCh:
				fmt.Printf("Error: %s", err)
			case <-interrupt:
				log.Println("interrupt received")
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
