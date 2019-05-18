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

// ProductStorage is an interface to fetch products from some persistent storage (mysql, etc)
// This is intended to normalize products (pairs) where some exchanges might list IOTA-USD as IOT-USD
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

// Product is a struct representing all the format of each product.
// Ex_ stands for exchange, He_ stands for helgart
type Product struct {
	Exchange string `json:"exchange"`
	ExPair   string `json:"ex_pair"`
	HePair   string `json:"he_pair"`
	ExBase   string `json:"ex_base"`
	ExQuote  string `json:"ex_quote"`
	HeBase   string `json:"he_base"`
	HeQuote  string `json:"he_quote"`
}

// ExProductMap is a map of pairs to product details
type ExProductMap map[string]Product

// ProductMap is a map that normalizes all products (pairs)
// First key is exchange, second key is pair
// TODO: make more descriptive
type ProductMap map[string]ExProductMap

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
		case KRAKEN:
			exPair = strings.Replace(product.ExPair, "-", "/", 1)
		case BITFINEX, BINANCE:
			exPair = strings.Replace(product.ExPair, "-", "", 1)
		default:
			exPair = product.ExPair
		}
		b.ProductMap[exchange][exPair] = product
	}
	return nil
}
