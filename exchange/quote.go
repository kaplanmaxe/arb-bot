package exchange

// Quote represents a quote from an exchange
type Quote struct {
	Exchange string
	Pair     string
	Price    string
}

// krakenPairMap normalizes the pair names for common use
var krakenPairMap = map[string]string{
	"XBT/USD": "BTC/USD",
}

// NewExchangeQuote returns a new exchange quote struct
func NewExchangeQuote(exchange, pair, price string) *Quote {
	p := pair
	if exchange == "kraken" && krakenPairMap[pair] != "" {
		p = krakenPairMap[pair]
	}
	return &Quote{
		Exchange: exchange,
		Pair:     p,
		Price:    price,
	}
}
