package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

// Client represents an API client
type Client struct {
	Pairs        []string
	quoteCh      chan<- broker.Quote
	errorCh      chan<- error
	api          api.Connector
	exchangeName string
}

// NewClient returns a new instance of the API
func NewClient(pairs []string, api api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) exchange.API {
	return &Client{
		Pairs:        pairs,
		quoteCh:      quoteCh,
		errorCh:      errorCh,
		api:          api,
		exchangeName: exchange.BINANCE,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context) {
	c.api.Connect(c.GetURL())
	c.StartTickerListener(ctx)
}

// FormatSubscribeRequest creates the type for a subscribe request
func (c *Client) FormatSubscribeRequest() interface{} {
	return nil
}

// StartTickerListener starts a new goroutine to listen for new ticker messages
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

// ParseTickerResponse parses the ticker response and returns a new instance of a broker.Quote
func (c *Client) ParseTickerResponse(msg []byte) (broker.Quote, error) {
	var err error
	var quote broker.Quote

	var res tickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return broker.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}
	if res.Pair != "" {
		quote = *broker.NewExchangeQuote(exchange.BINANCE, res.Pair, res.Price)
	}
	return quote, nil
}

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/bnbbtc@ticker"}
}
