package coinbase

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

type Client struct {
	Pairs []string
}

func NewClient(pairs []string, quoteCh chan<- broker.Quote) exchange.API {
	return &Client{
		Pairs: pairs,
	}
}

func (c *Client) FormatSubscribeRequest() interface{} {
	return &subscribeRequest{
		Type:       "subscribe",
		ProductIDs: c.Pairs,
		Channels: []struct {
			Name       string   `json:"name"`
			ProductIDs []string `json:"product_ids"`
		}{
			{
				Name:       "ticker",
				ProductIDs: c.Pairs,
			},
		},
	}
}

func (c *Client) ParseTickerResponse(msg []byte) broker.Quote {
	var err error
	var quote broker.Quote

	var res tickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		log.Fatal("Unmarshal", err)
	}
	if res.Pair != "" {
		quote = *broker.NewExchangeQuote("coinbase", res.Pair, res.Price)
	}
	return quote
}

func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "ws-feed.pro.coinbase.com"}
}
