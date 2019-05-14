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
	doneCh := make(chan struct{}, 1)
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error)

	signal.Notify(interrupt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())

	// engine := exchange.NewEngine([]exchange.API{
	// 	kraken.NewClient(api.NewSource(exchange.KRAKEN), quoteCh, errorCh),
	// 	coinbase.NewClient(api.NewSource(exchange.COINBASE), quoteCh, errorCh),
	// 	binance.NewClient(api.NewSource(exchange.BINANCE), quoteCh, errorCh),
	// })
	// engine.Start(ctx)

	kr := kraken.NewClient(api.NewSource(exchange.KRAKEN), quoteCh, errorCh)
	cb := coinbase.NewClient(api.NewSource(exchange.COINBASE), quoteCh, errorCh)
	bin := binance.NewClient(api.NewSource(exchange.BINANCE), quoteCh, errorCh)

	err := kr.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = cb.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = bin.Start(ctx)
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
