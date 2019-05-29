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

// side represents an order book side for the below enum
type side int

const (
	// BIDS represents the bids side of the book
	BIDS = iota
	// ASKS represemts the asks side of the book
	ASKS
)

// Rates map represents a map of pairs we will want to triangulate.
// Ex: "BTC-USD": 8000
type ratesMap map[string]float64

// // cryptoRates represents a map of commonly used crypto quote currencies and the exchange rates
// // back to a given fiat
// var cryptoRates ratesMap

// fiatQuoteCurrencies represents a slice of fiat currences we will value pairs in
var fiatQuoteCurrencies = []string{"USD", "EUR", "GBP", "CAD"}

// cryptoQuoteCurrencies is a slice of currencies that we want to triangulate
var cryptoQuoteCurrencies = []string{"BTC", "ETH", "BNB"}

// stableCoins is a slice of stable coins that will be used as quote currencies
var stableCoins = []string{"USDT", "USDC", "TUSD", "PAX"}

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
	HeBase   string  `json:"he_base"`
	HeQuote  string  `json:"he_quote"`
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
}

// ArbMarket represents a market for arbitrage with a spread, low exchange, and high exchange
type ArbMarket struct {
	HeBase string     `json:"he_base"`
	Spread float64    `json:"spread"`
	Low    MarketSide `json:"low"`
	High   MarketSide `json:"high"`
}

// NewArbMarket returns a new ArbMarket
func NewArbMarket(heBase string, low, high MarketSide) *ArbMarket {
	spread := ((high.TriangulatedPrice - low.TriangulatedPrice) / low.TriangulatedPrice) * 100
	return &ArbMarket{
		HeBase: heBase,
		Spread: spread,
		Low:    low,
		High:   high,
	}
}

// Market represents a market in the ActiveMarketMap for a given pair
type Market struct {
	Bids []MarketSide `json:"bids"`
	Asks []MarketSide `json:"asks"`
}

// MarketSide is a quote from an exchange
type MarketSide struct {
	Exchange          string  `json:"exchange"`
	Price             float64 `json:"price"`
	TriangulatedPrice float64 `json:"triangulated_price,omitempty"`
	ExPair            string  `json:"ex_pair"`
	HePair            string  `json:"he_pair"`
}

// ActiveMarketMap is a map representing markets broker is pulling from and receiving data from
// the key will be the HePair and the value will be a Quote
// 	{
// 		"REP": {
//  		"Bids": [
//  			{"exchange": "Coinbase", "price": 123, "ExPair": "REP-USD"},
//  			{"exchange": "Kraken", "price": 456, "ExPair": "REP-USDT"},
//  		],
//  		"Asks": [
//  			{"exchange": "Coinbase", "price": 123, "ExPair": "REP-USD"},
//  			{"exchange": "Kraken", "price": 456, "ExPair": "REP-USDC"},
//  		]}
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
	cryptoRates   ratesMap
	forexRates    ratesMap
}

// NewBroker returns a new broker interface
func NewBroker(exchanges []Exchange, db ProductStorage) *Broker {
	// Forex rates to value things back to USD
	// TODO: fetch programatically
	var forexRates = ratesMap{
		"GBP-USD": 1.27,
		"CAD-USD": 0.74,
		"EUR-USD": 1.12,
		"USD-USD": 1,
	}
	return &Broker{
		exchanges:     exchanges,
		db:            db,
		ProductMap:    make(ProductMap),
		ActiveMarkets: make(ActiveMarketMap),
		cryptoRates:   make(ratesMap),
		forexRates:    forexRates,
	}
}

// indexOf returns the index of the needle in the haystack
// if not found it returns -1 similarily to javascript
// implementation
func indexOf(needle string, haystack []string) int {
	index := -1
	for key, val := range haystack {
		if needle == val {
			index = key
			break
		}
	}
	return index
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
	var m MarketSide
	if side == BIDS {
		for key, val := range b.ActiveMarkets[market.HeBase].Bids {
			if market.Exchange == val.Exchange && market.HePair == val.HePair {

				b.ActiveMarkets[market.HeBase].Bids[key].Price = market.Bid
				b.calculatedTriangulatedPrice(&b.ActiveMarkets[market.HeBase].Bids[key], market.HeBase, market.HeQuote)
				return
			}
		}
		m = b.mapMarketSide(BIDS, market)
		if m.Price > 0 {
			b.ActiveMarkets[market.HeBase].Bids = append(b.ActiveMarkets[market.HeBase].Bids, m)
		}

	} else if side == ASKS {
		for key, val := range b.ActiveMarkets[market.HeBase].Asks {
			if market.Exchange == val.Exchange && market.HePair == val.HePair {
				b.ActiveMarkets[market.HeBase].Asks[key].Price = market.Ask
				b.calculatedTriangulatedPrice(&b.ActiveMarkets[market.HeBase].Asks[key], market.HeBase, market.HeQuote)
				return
			}
		}
		m = b.mapMarketSide(ASKS, market)
		if m.Price > 0 {
			b.ActiveMarkets[market.HeBase].Asks = append(b.ActiveMarkets[market.HeBase].Asks, m)
		}

	}
}

func (b *Broker) mapMarketSide(s side, market *ActiveMarket) MarketSide {
	var price float64
	if s == BIDS {
		price = market.Bid
	} else if s == ASKS {
		price = market.Ask
	}
	m := MarketSide{
		Exchange: market.Exchange,
		Price:    price,
		ExPair:   market.ExPair,
		HePair:   market.HePair,
	}
	b.calculatedTriangulatedPrice(&m, market.HeBase, market.HeQuote)
	return m
}

func (b *Broker) calculatedTriangulatedPrice(m *MarketSide, base, quote string) {
	forexIndex := indexOf(quote, fiatQuoteCurrencies)
	if forexIndex > -1 {
		// TODO: allow to value in more fiats
		index := fmt.Sprintf("%s-USD", quote)
		m.TriangulatedPrice = m.Price * b.forexRates[index]
		return
	}

	cryptoIndex := indexOf(quote, cryptoQuoteCurrencies)
	if cryptoIndex > -1 {
		// TODO: allow to value in more fiats
		index := fmt.Sprintf("%s-USD", quote)
		if b.cryptoRates[index] > 0 {
			m.TriangulatedPrice = m.Price * b.cryptoRates[index]
			return
		}
	}

	stableCoinIndex := indexOf(quote, stableCoins)
	if stableCoinIndex > -1 {
		m.TriangulatedPrice = m.Price
		return
	}

	// if there is no triangulation, remove the price to remove the market altogether
	m.Price = 0
	// m.TriangulatedPrice = m.Price
}

// InsertActiveMarket inserts a quote into the ActiveMarketMap
func (b *Broker) InsertActiveMarket(market *ActiveMarket) {
	// We check if the pair is already in the active market map
	if _, ok := b.ActiveMarkets[market.HeBase]; !ok {
		// If not intialize slice
		b.ActiveMarkets[market.HeBase] = &Market{}
	}

	// if the exchange is coinbase and the market is a crypto quote currency we use and a fiat pair
	// we'll store the rate for triangulation
	fiatIndex := indexOf(market.HeQuote, fiatQuoteCurrencies)
	cryptoIndex := indexOf(market.HeBase, cryptoQuoteCurrencies)
	if market.Exchange == "coinbase" && (fiatIndex > -1 && cryptoIndex > -1) {
		// TODO: should we always be using ask here?
		b.cryptoRates[market.HePair] = market.Ask
	} else if market.Exchange == "binance" && market.HeBase == "BNB" && market.HeQuote == "USDT" {
		// TODO: actually pull this
		b.cryptoRates[market.HePair] = market.Ask
	}
	b.insertMarketIntoMarketSide(BIDS, market)
	b.insertMarketIntoMarketSide(ASKS, market)

	// We sort here to easily find the high and low price
	sort.Slice(b.ActiveMarkets[market.HeBase].Bids, func(i, j int) bool {
		return b.ActiveMarkets[market.HeBase].Bids[i].TriangulatedPrice > b.ActiveMarkets[market.HeBase].Bids[j].TriangulatedPrice
	})

	sort.Slice(b.ActiveMarkets[market.HeBase].Asks, func(i, j int) bool {
		return b.ActiveMarkets[market.HeBase].Asks[i].TriangulatedPrice < b.ActiveMarkets[market.HeBase].Asks[j].TriangulatedPrice
	})
}
