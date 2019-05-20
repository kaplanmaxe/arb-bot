package exchange

// Quote represents a quote from an exchange
type Quote struct {
	Exchange   string
	ExPair     string
	HePair     string
	ExBase     string
	ExQuote    string
	HeBase     string
	HeQuote    string
	Price      string
	PriceFloat float64 // price represented as a float
}

// krakenPairMap normalizes the pair names for common use
var krakenPairMap = map[string]string{
	"XBT/USD": "BTC/USD",
}

// NewExchangeQuote returns a new exchange quote struct
func NewExchangeQuote(exchange, pair, price string) *Quote {
	p := pair
	// if exchange == "kraken" && krakenPairMap[pair] != "" {
	// 	p = krakenPairMap[pair]
	// }
	return &Quote{
		Exchange: exchange,
		HePair:   p,
		Price:    price,
	}
}
