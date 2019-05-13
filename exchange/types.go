package exchange

type CoinbaseSubscribeRequest struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	} `json:"channels"`
}

type CoinbaseTickerResponse struct {
	Pair  string `json:"product_id"`
	Price string `json:"best_ask"`
}
