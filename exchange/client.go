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
	Start(context.Context)
	GetURL() *url.URL
	ParseTickerResponse([]byte) ([]broker.Quote, error)
	FormatSubscribeRequest() interface{}
}

type Engine interface {
	Start(context.Context)
}

type Group struct {
	exchanges []API
}

func NewEngine(exchanges []API) Engine {
	return &Group{
		exchanges: exchanges,
	}
}

func (g *Group) Start(ctx context.Context) {
	for _, exchange := range g.exchanges {
		exchange.Start(ctx)
	}
}
