package binance

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

type Client struct {
	Pairs   []string
	quoteCh chan<- broker.Quote
	api     api.Connector
}

func NewClient(pairs []string, api api.Connector, quoteCh chan<- broker.Quote) exchange.API {
	return &Client{
		Pairs:   pairs,
		quoteCh: quoteCh,
		api:     api,
	}
}

func (c *Client) Start(ctx context.Context) {
	c.api.Connect(c.GetURL())
	c.StartTickerListener(ctx)
}

func (c *Client) FormatSubscribeRequest() interface{} {
	return nil
}

func (c *Client) StartTickerListener(ctx context.Context) {
	go func() {
	cLoop:
		for {
			message, err := c.api.ReadMessage()
			if err != nil {
				// TODO: fix
				log.Println("cb read2:", err, message)
				return
			}

			select {
			case <-ctx.Done():
				err := c.api.Close()
				if err != nil {
					log.Printf("Error closing %s: %s", exchange.COINBASE, err)
				}
				break cLoop
			default:
				c.quoteCh <- c.ParseTickerResponse(message)
			}
		}
	}()
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
		quote = *broker.NewExchangeQuote(exchange.BINANCE, res.Pair, res.Price)
	}
	return quote
}

func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/bnbbtc@ticker"}
}
