package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/bitfinex"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

func main() {
	log.Print("Starting quote server")
	interrupt := make(chan os.Signal, 1)
	doneCh := make(chan struct{}, 1)
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error)

	signal.Notify(interrupt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	broker := exchange.NewBroker([]exchange.API{
		// kraken.NewClient(api.NewSource(exchange.KRAKEN), quoteCh, errorCh),
		// coinbase.NewClient(api.NewSource(exchange.COINBASE), quoteCh, errorCh),
		// binance.NewClient(api.NewSource(exchange.BINANCE), quoteCh, errorCh),
		bitfinex.NewClient(api.NewSource(exchange.BITFINEX), quoteCh, errorCh),
	})

	err := broker.Start(ctx)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case quote := <-quoteCh:
				log.Printf("Quote: %#v", quote)
			case err := <-errorCh:
				fmt.Printf("Error: %s\n", err)
			case <-interrupt:
				log.Println("interrupt received")
				cancel()
				select {
				// TODO: race condition. fix
				case <-time.After(7 * time.Second):
					close(doneCh)
					return
				}
			default:
			}
		}
	}()
	<-doneCh
}
