package exchange

// Quote represents a quote from an exchange
type Quote struct {
	Exchange string `json:"exchange"`
	ExPair   string `json:"ex_pair"`
	HePair   string `json:"he_pair"`
	ExBase   string `json:"ex_base"`
	ExQuote  string `json:"ex_quote"`
	HeBase   string `json:"he_base"`
	HeQuote  string `json:"he_quote"`
	// Price      string  `json:"price"` // TODO: remove
	Bid        string  `json:"bid"`
	Ask        string  `json:"ask"`
	PriceFloat float64 `json:"price_float,omitempty"` // price represented as a float
}

// NewExchangeQuote returns a new exchange quote struct
func NewExchangeQuote(exchange, pair, bid, ask string) *Quote {
	p := pair
	return &Quote{
		Exchange: exchange,
		HePair:   p,
		Bid:      bid,
		Ask:      ask,
	}
}
