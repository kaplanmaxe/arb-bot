package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kaplanmaxe/helgart/broker/api"
	"github.com/kaplanmaxe/helgart/broker/exchange"
)

// Client represents an API client
type Client struct {
	pairs        []string
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
		exchangeName: exchange.COINBASE,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context, productMap exchange.ProductMap) error {
	c.productMap = productMap[c.exchangeName]
	err := c.GetPairs()
	if err != nil {
		return err
	}
	err = c.API.Connect(c.GetURL())
	if err != nil {
		return err
	}
	err = c.API.SendSubscribeRequest(c.FormatSubscribeRequest())
	if err != nil {
		return err
	}
	go c.StartTickerListener(ctx)
	return nil
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
			break cLoop
		default:
			res, err := c.ParseTickerResponse(message)
			if err != nil {
				c.errorCh <- err
			} else if len(res) > 0 {
				if res[0].HePair != "" {
					c.quoteCh <- res[0]
				}
			}
		}
	}
}

// ParseTickerResponse parses the ticker response and returns a new instance of a exchange.Quote
func (c *Client) ParseTickerResponse(msg []byte) ([]exchange.Quote, error) {
	var err error
	var quotes []exchange.Quote

	var res TickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return []exchange.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}
	if res.Pair != "" {
		product := c.productMap[res.Pair]
		quotes = append(quotes, exchange.Quote{
			Exchange: c.exchangeName,
			Price:    res.Price,
			ExPair:   product.ExPair,
			HePair:   product.HePair,
			ExBase:   product.ExBase,
			HeBase:   product.HeBase,
			ExQuote:  product.ExQuote,
			HeQuote:  product.HeQuote,
		})
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
