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

const (
	// BIDS represents the bids side of the book
	BIDS = iota
	// ASKS represemts the asks side of the book
	ASKS
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
	Exchange string `json:"exchange"`
	HePair   string `json:"he_pair"`
	ExPair   string `json:"ex_pair"`
	// Price    float64 `json:"price"`
	Bid float64 `json:"bid"`
	Ask float64 `json:"ask"`
}

// ArbMarket represents a market for arbitrage with a spread, low exchange, and high exchange
type ArbMarket struct {
	HePair string     `json:"he_pair"`
	Spread float64    `json:"spread"`
	Low    MarketSide `json:"low"`
	High   MarketSide `json:"high"`
}

// NewArbMarket returns a new ArbMarket
func NewArbMarket(hePair string, low, high MarketSide) *ArbMarket {
	spread := ((high.Price - low.Price) / low.Price) * 100
	return &ArbMarket{
		HePair: hePair,
		Spread: spread,
		Low:    low,
		High:   high,
	}
}

// Market represents a market in the ActiveMarketMap for a given pair
type Market struct {
	HePair string       `json:"he_pair"`
	Bids   []MarketSide `json:"bids"`
	Asks   []MarketSide `json:"asks"`
}

// MarketSide is a quote from an exchange
type MarketSide struct {
	Exchange string  `json:"exchange"`
	Price    float64 `json:"price"`
	ExPair   string  `json:"ex_pair"`
}

// ActiveMarketMap is a map representing markets broker is pulling from and receiving data from
// the key will be the HePair and the value will be a Quote
// 	{
// 		"REP-USD": {
// 		"HePair": "REP-USD",
// 		"Bids": [
// 			{"exchange": "Coinbase", "price": 123, "ExPair": "REP-USD"},
// 			{"exchange": "Kraken", "price": 456, "ExPair": "REP-USD"},
// 		],
// 		"Asks": [
// 			{"exchange": "Coinbase", "price": 123, "ExPair": "REP-USD"},
// 			{"exchange": "Kraken", "price": 456, "ExPair": "REP-USD"},
// 		]}
// 	}
type ActiveMarketMap map[string]*Market

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
func (b *Broker) Start(ctx context.Context, doneCh chan<- struct{}) error {
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
		err := exchange.Start(ctx, b.ProductMap, doneCh)
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

func (b *Broker) insertMarketIntoMarketSide(side int, market *ActiveMarket) {
	if side == BIDS {
		for key, val := range b.ActiveMarkets[market.HePair].Bids {
			if market.Exchange == val.Exchange {
				b.ActiveMarkets[market.HePair].Bids[key].Price = market.Bid
				return
			}
		}
		b.ActiveMarkets[market.HePair].Bids = append(b.ActiveMarkets[market.HePair].Bids, MarketSide{
			Exchange: market.Exchange,
			Price:    market.Bid,
			ExPair:   market.ExPair,
		})
	} else if side == ASKS {
		for key, val := range b.ActiveMarkets[market.HePair].Asks {
			if market.Exchange == val.Exchange {
				b.ActiveMarkets[market.HePair].Asks[key].Price = market.Ask
				return
			}
		}
		b.ActiveMarkets[market.HePair].Asks = append(b.ActiveMarkets[market.HePair].Asks, MarketSide{
			Exchange: market.Exchange,
			Price:    market.Ask,
			ExPair:   market.ExPair,
		})
	}
}

// InsertActiveMarket inserts a quote into the ActiveMarketMap
func (b *Broker) InsertActiveMarket(market *ActiveMarket) {
	// We check if the pair is already in the active market map
	if _, ok := b.ActiveMarkets[market.HePair]; !ok {
		// If not intialize slice
		b.ActiveMarkets[market.HePair] = &Market{
			HePair: market.HePair,
		}
	}
	b.insertMarketIntoMarketSide(BIDS, market)
	b.insertMarketIntoMarketSide(ASKS, market)

	// We sort here to easily find the high and low price
	sort.Slice(b.ActiveMarkets[market.HePair].Bids, func(i, j int) bool {
		return b.ActiveMarkets[market.HePair].Bids[i].Price > b.ActiveMarkets[market.HePair].Bids[j].Price
	})

	sort.Slice(b.ActiveMarkets[market.HePair].Asks, func(i, j int) bool {
		return b.ActiveMarkets[market.HePair].Asks[i].Price < b.ActiveMarkets[market.HePair].Asks[j].Price
	})
}
