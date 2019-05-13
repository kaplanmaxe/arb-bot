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
	ParseTickerResponse([]byte) (broker.Quote, error)
	FormatSubscribeRequest() interface{}
}

// Connector is an interface containing methods to perform various actions
// on ticker websocket connections
type Connector interface {
	Start(context.Context)
	Connect() error
	readMessage() ([]byte, error)
	SendSubscribeRequest(interface{}) error
	SendSubscribeRequestWithResponse(context.Context, interface{}) ([]byte, error)
	writeMessage([]byte) error
	StartTickerListener(context.Context)
	Close() error
}
