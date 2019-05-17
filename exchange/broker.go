package exchange

import (
	"context"

	"github.com/kaplanmaxe/helgart/api"
)

const (
	// KRAKEN represents the kraken api
	KRAKEN = "kraken"
	// COINBASE represents the coinbase api
	COINBASE = "coinbase"
	// BINANCE represents the binance api
	BINANCE = "binance"
	// BITFINEX represents the bitfinex api
	BITFINEX = "bitfinex"
)

// NewBroker returns a new broker interface
func NewBroker(exchanges []api.Exchange) Broker {
	return &broker{
		exchanges: exchanges,
	}
}

// Broker is an interface to start a new instance of a broker
type Broker interface {
	Start(context.Context) error
}

// Group is a struct representing a group of exchanges
type broker struct {
	exchanges []api.Exchange
}

// Start starts a new exchange engine
func (b *broker) Start(ctx context.Context) error {
	for _, exchange := range b.exchanges {
		err := exchange.Start(ctx)
		if err != nil {
			return err
		}

	}
	return nil
}
