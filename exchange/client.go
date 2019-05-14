package exchange

import (
	"context"
	"net/url"

	"github.com/kaplanmaxe/helgart/broker"
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

// API is an interface each exchange client should satisfy
type API interface {
	Start(context.Context) error
	GetURL() *url.URL
	ParseTickerResponse([]byte) ([]broker.Quote, error)
	FormatSubscribeRequest() interface{}
}

// Engine is an interface for a new exchange engine
type Engine interface {
	Start(context.Context)
}

// Group is a struct representing a group of exchanges
type Group struct {
	exchanges []API
}

// NewEngine returns an exchange engine
func NewEngine(exchanges []API) Engine {
	return &Group{
		exchanges: exchanges,
	}
}

// Start starts a new exchange engine
func (g *Group) Start(ctx context.Context) {
	for _, exchange := range g.exchanges {
		exchange.Start(ctx)
	}
}
