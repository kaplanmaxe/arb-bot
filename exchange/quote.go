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

type Product struct {
	Exchange string `json:"exchange"`
	ExPair   string `json:"ex_pair"`
	HePair   string `json:"he_pair"`
	ExBase   string `json:"ex_base"`
	ExQuote  string `json:"ex_quote"`
	HeBase   string `json:"he_base"`
	HeQuote  string `json:"he_quote"`
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
