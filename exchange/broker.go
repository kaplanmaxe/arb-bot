package exchange

import (
	"context"
	"fmt"
	"strings"
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

type ProductStorage interface {
	Connect() error
	FetchProducts() ([]Product, error)
}

// NewBroker returns a new broker interface
func NewBroker(exchanges []Exchange, db ProductStorage) *Broker {
	return &Broker{
		exchanges:  exchanges,
		db:         db,
		ProductMap: make(ProductMap),
	}
}

type Product struct {
	Exchange string `json:"exchange"`
	ExPair   string `json:"ex_pair"`
	HePair   string `json:"he_pair"`
	ExBase   string `json:"ex_base"`
	ExQuote  string `json:"ex_quote"`
	HeBase   string `json:"he_base"`
	HeQuote  string `json:"he_quote"`
}

// ExchangeProductMap is a map of pairs to product details
type ExchangeProductMap map[string]Product

// ProductMap is a map that normalizes all products (pairs)
// First key is exchange, second key is pair
// TODO: make more descriptive
type ProductMap map[string]ExchangeProductMap

// Broker is an interface to start a new instance of a broker
// type Broker interface {
// 	Start(context.Context) error
// 	buildProductMap() error
// }

// Broker is a struct representing a group of exchanges
type Broker struct {
	exchanges  []Exchange
	db         ProductStorage
	ProductMap ProductMap
}

// Start starts a new exchange engine
func (b *Broker) Start(ctx context.Context) error {
	err := b.buildProductMap()
	if err != nil {
		return fmt.Errorf("Error fetching product map: %s", err)
	}
	for _, exchange := range b.exchanges {
		err := exchange.Start(ctx, b.ProductMap)
		if err != nil {
			return err
		}

	}
	return nil
}

func (b *Broker) buildProductMap() error {
	products, err := b.db.FetchProducts()
	if err != nil {
		return fmt.Errorf("Error fetching products: %s", err)
	}
	for _, product := range products {
		exchange := strings.ToLower(product.Exchange)
		if len(b.ProductMap[exchange]) == 0 {
			b.ProductMap[exchange] = map[string]Product{}
		}
		var exPair string
		switch exchange {
		case "kraken":
			exPair = strings.Replace(product.ExPair, "-", "/", 1)
		default:
			exPair = product.ExPair
		}
		b.ProductMap[exchange][exPair] = product
	}
	return nil
}
