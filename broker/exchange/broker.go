package exchange

import (
	"context"
	"fmt"
	"sort"
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
	FetchAllProducts() ([]Product, error)
	FetchArbProducts() (ArbProductMap, error)
}

// ProductCache is an interface for an in-memory cache instance (Ex: Redis)
type ProductCache interface {
	Connect() error
	SetPair(string, string, string) error
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
type ProductMap map[string]ExProductMap

// ArbProductMap represents a map of all pairs that have more than one market
type ArbProductMap map[string]struct{}

// ActiveMarket represents a market to go into the ActiveMarketMap
type ActiveMarket struct {
	Exchange string  `json:"exchange"`
	HePair   string  `json:"he_pair"`
	ExPair   string  `json:"ex_pair"`
	Price    float64 `json:"price"`
}

// ArbMarket represents a market for arbitrage with a spread, low exchange, and high exchange
type ArbMarket struct {
	HePair string       `json:"he_pair"`
	Spread float64      `json:"spread"`
	Low    ActiveMarket `json:"low"`
	High   ActiveMarket `json:"high"`
}

// NewArbMarket returns a new ArbMarket
func NewArbMarket(hePair string, low, high ActiveMarket) *ArbMarket {
	spread := ((high.Price - low.Price) / low.Price) * 100
	return &ArbMarket{
		HePair: hePair,
		Spread: spread,
		Low:    low,
		High:   high,
	}
}

// ActiveMarketMap is a map representing markets broker is pulling from and receiving data from
// the key will be the HePair and the value will be a Quote
type ActiveMarketMap map[string][]ActiveMarket

// Broker is a struct representing a group of exchanges
type Broker struct {
	ProductMap    ProductMap
	ArbProducts   ArbProductMap
	ActiveMarkets ActiveMarketMap
	exchanges     []Exchange
	db            ProductStorage
	cache         ProductCache
}

// NewBroker returns a new broker interface
func NewBroker(exchanges []Exchange, db ProductStorage) *Broker {
	return &Broker{
		exchanges:     exchanges,
		db:            db,
		ProductMap:    make(ProductMap),
		ActiveMarkets: make(ActiveMarketMap),
	}
}

// Start starts a new exchange engine
func (b *Broker) Start(ctx context.Context) error {
	err := b.buildProductMap()
	if err != nil {
		return fmt.Errorf("Error fetching product map: %s", err)
	}
	arbProducts, err := b.db.FetchArbProducts()
	if err != nil {
		return fmt.Errorf("Error fetching arb pairs: %s", err)
	}
	b.ArbProducts = arbProducts
	for _, exchange := range b.exchanges {
		err := exchange.Start(ctx, b.ProductMap)
		if err != nil {
			return err
		}

	}
	return nil
}

func (b *Broker) buildProductMap() error {
	products, err := b.db.FetchAllProducts()
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

// InsertActiveMarket inserts a quote into the ActiveMarketMap
func (b *Broker) InsertActiveMarket(market *ActiveMarket) {
	// We check if the pair is already in the active market map
	if _, ok := b.ActiveMarkets[market.HePair]; !ok {
		// If not intialize slice
		b.ActiveMarkets[market.HePair] = []ActiveMarket{}
	}
	// We now check if we have an exchange price in the active market map for the given pair
	// and get the key in the slice
	exchangeKey := -1
	for key, val := range b.ActiveMarkets[market.HePair] {
		if val.Exchange == market.Exchange {
			exchangeKey = key
		}
	}
	// If exchange price isn't in the active market map for the exchange, add it
	if exchangeKey == -1 {
		b.ActiveMarkets[market.HePair] = append(b.ActiveMarkets[market.HePair], *market)
	} else {
		// If it's already there, simply update the price
		b.ActiveMarkets[market.HePair][exchangeKey].Price = market.Price
	}

	// We sort here to easily find the high and low price
	sort.Slice(b.ActiveMarkets[market.HePair], func(i, j int) bool {
		return b.ActiveMarkets[market.HePair][i].Price > b.ActiveMarkets[market.HePair][j].Price
	})
}
