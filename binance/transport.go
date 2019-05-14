package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
				} else if len(res) > 0 {
					for _, val := range res {
						c.quoteCh <- val
					}
				}
			}
		}
	}()
}

// ParseTickerResponse parses the ticker response and returns a new instance of a broker.Quote
func (c *Client) ParseTickerResponse(msg []byte) ([]broker.Quote, error) {
	var err error
	var quotes []broker.Quote

	var res []tickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return []broker.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}

	for _, val := range res {
		if val.Pair != "" {
			quotes = append(quotes, *broker.NewExchangeQuote(exchange.BINANCE, val.Pair, val.Price))
		}
	}
	return quotes, nil
}

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/!ticker@arr"}
}

// GetPairs returns all pairs for an exchange
func (c *Client) GetPairs() error {
	u := url.URL{Scheme: "https", Host: "api.binance.com", Path: "/api/v1/exchangeInfo"}
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var response productsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	var pairs []string
	for _, val := range response.Symbols {
		pairs = append(pairs, val.Pair)
	}
	c.Pairs = pairs
	return nil
}
