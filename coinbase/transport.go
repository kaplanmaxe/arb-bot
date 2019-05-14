package coinbase

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
	pairs        []string
	quoteCh      chan<- broker.Quote
	errorCh      chan<- error
	api          api.Connector
	exchangeName string
}

// NewClient returns a new instance of the API
func NewClient(api api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) exchange.API {
	return &Client{
		quoteCh:      quoteCh,
		errorCh:      errorCh,
		api:          api,
		exchangeName: exchange.COINBASE,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context) {
	c.GetPairs()
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

// FormatSubscribeRequest creates the type for a subscribe request
func (c *Client) FormatSubscribeRequest() interface{} {
	return &subscribeRequest{
		Type:       "subscribe",
		ProductIDs: c.pairs,
		Channels: []struct {
			Name       string   `json:"name"`
			ProductIDs []string `json:"product_ids"`
		}{
			{
				Name:       "ticker",
				ProductIDs: c.pairs,
			},
		},
	}
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
					if res[0].Pair != "" {
						c.quoteCh <- res[0]
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

	var res tickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return []broker.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}
	if res.Pair != "" {
		quotes = append(quotes, *broker.NewExchangeQuote(exchange.COINBASE, res.Pair, res.Price))
	}
	return quotes, nil
}

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "ws-feed.pro.coinbase.com"}
}

// GetPairs returns all pairs for an exchange
func (c *Client) GetPairs() error {
	u := url.URL{Scheme: "https", Host: "api.pro.coinbase.com", Path: "/products"}
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var response []productsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	var pairs []string
	for _, val := range response {
		pairs = append(pairs, val.Pair)
	}
	c.pairs = pairs
	return nil
}
