package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/kaplanmaxe/helgart/broker/api"
	"github.com/kaplanmaxe/helgart/broker/exchange"
)

// Client represents an API client
type Client struct {
	quoteCh      chan<- exchange.Quote
	errorCh      chan<- error
	API          api.WebSocketHelper
	exchangeName string
	productMap   exchange.ExProductMap
}

// NewClient returns a new instance of the API
func NewClient(api api.WebSocketHelper, quoteCh chan<- exchange.Quote, errorCh chan<- error) *Client {
	return &Client{
		quoteCh:      quoteCh,
		errorCh:      errorCh,
		API:          api,
		exchangeName: exchange.BINANCE,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context, productMap exchange.ProductMap, doneCh chan<- struct{}) error {
	c.productMap = productMap[c.exchangeName]
	err := c.API.Connect(c.GetURL())
	if err != nil {
		return err
	}
	go c.StartTickerListener(ctx, doneCh)
	return nil
}

// StartTickerListener starts a new goroutine to listen for new ticker messages
func (c *Client) StartTickerListener(ctx context.Context, doneCh chan<- struct{}) {
cLoop:
	for {
		message, err := c.API.ReadMessage()
		if err != nil {
			c.errorCh <- fmt.Errorf("Error reading from %s: %s", c.exchangeName, err)
			return
		}

		select {
		case <-ctx.Done():
			err := c.API.Close()
			if err != nil {
				c.errorCh <- fmt.Errorf("Error closing %s: %s", c.exchangeName, err)
			}
			doneCh <- struct{}{}
			break cLoop
		default:
			res, err := c.ParseTickerResponse(message)
			if err != nil {
				c.errorCh <- err
			} else if len(res) > 0 {
				for _, val := range res {
					if val.HePair != "" {
						c.quoteCh <- val
					}
				}
			}
		}
	}
}

// ParseTickerResponse parses the ticker response and returns a new instance of a exchange.Quote
func (c *Client) ParseTickerResponse(msg []byte) ([]exchange.Quote, error) {
	var err error
	var quotes []exchange.Quote

	var res []TickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return []exchange.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}

	for _, val := range res {

		if val.Pair != "" {
			product := c.productMap[val.Pair]
			quotes = append(quotes, exchange.Quote{
				Exchange: c.exchangeName,
				Price:    val.Price,
				ExPair:   product.ExPair,
				HePair:   product.HePair,
				ExBase:   product.ExBase,
				HeBase:   product.HeBase,
				ExQuote:  product.ExQuote,
				HeQuote:  product.HeQuote,
			})
		}
	}
	return quotes, nil
}

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/!ticker@arr"}
}
