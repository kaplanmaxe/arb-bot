package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

type Client struct {
	Pairs        []string
	quoteCh      chan<- broker.Quote
	errorCh      chan<- error
	api          api.Connector
	exchangeName string
}

func NewClient(pairs []string, api api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) exchange.API {
	return &Client{
		Pairs:        pairs,
		quoteCh:      quoteCh,
		errorCh:      errorCh,
		api:          api,
		exchangeName: exchange.COINBASE,
	}
}

func (c *Client) Start(ctx context.Context) {
	c.api.Connect(c.GetURL())
	err := c.api.SendSubscribeRequest(c.FormatSubscribeRequest())
	if err != nil {
		go func() {
			c.errorCh <- err
		}()
		return
	}
	c.StartTickerListener(ctx)
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

func (c *Client) StartTickerListener(ctx context.Context) {
	go func() {
	cLoop:
		for {
			message, err := c.api.ReadMessage()
			if err != nil {
				c.errorCh <- fmt.Errorf("Error reading from %s: %s", c.exchangeName, err)
				return
			}

			select {
			case <-ctx.Done():
				err := c.api.Close()
				if err != nil {
					c.errorCh <- fmt.Errorf("Error closing %s: %s", c.exchangeName, err)
				}
				break cLoop
			default:
				res, err := c.ParseTickerResponse(message)
				if err != nil {
					c.errorCh <- err
				} else if res.Pair != "" {
					c.quoteCh <- res
				}
			}
		}
	}()
}

func (c *Client) ParseTickerResponse(msg []byte) (broker.Quote, error) {
	var err error
	var quote broker.Quote

	var res tickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return broker.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}
	if res.Pair != "" {
		quote = *broker.NewExchangeQuote(exchange.COINBASE, res.Pair, res.Price)
	}
	return quote, nil
}

func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "ws-feed.pro.coinbase.com"}
}
