package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

// Client represents an API client
type Client struct {
	Pairs          []string
	quoteCh        chan<- broker.Quote
	errorCh        chan<- error
	api            api.Connector
	channelPairMap ChannelPairMap
	exchangeName   string
}

// NewClient returns a new instance of the API
func NewClient(pairs []string, api api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) exchange.API {
	return &Client{
		Pairs:          pairs,
		quoteCh:        quoteCh,
		errorCh:        errorCh,
		api:            api,
		channelPairMap: make(ChannelPairMap),
		exchangeName:   exchange.KRAKEN,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context) {
	c.api.Connect(c.GetURL())
	var wg sync.WaitGroup
	wg.Add(1)
	err := c.SendSubscribeRequest(&wg, c.FormatSubscribeRequest())
	if err != nil {
		go func() {
			c.errorCh <- err
		}()
		return
	}
	wg.Wait()
	c.StartTickerListener(ctx)
}

// SendSubscribeRequest overrides the interface method and sends a subscription request and listens
// for a response
func (c *Client) SendSubscribeRequest(wg *sync.WaitGroup, req interface{}) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Error marshalling %s subscribe request: %s", c.exchangeName, err)
	}
	err = c.api.WriteMessage(payload)
	if err != nil {
		return fmt.Errorf("Error sending subscribe request for %s: %s", c.exchangeName, err)
	}
	go func() {
		var mtx sync.Mutex
		subs := 0
	Loop:
		for {
			message, err := c.api.ReadMessage()
			if err != nil {
				c.errorCh <- fmt.Errorf("Error reading message from %s: %s", c.exchangeName, err)
				return
			}

			var subStatusResponse subscriptionResponse
			err = json.Unmarshal(message, &subStatusResponse)
			if err != nil {
				c.errorCh <- fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
			}
			if subs < len(c.Pairs) {
				c.channelPairMap[subStatusResponse.ChannelID] = subStatusResponse.Pair
			} else {
				wg.Done()
				break Loop
			}
			mtx.Lock()
			subs++
			mtx.Unlock()
		}
		return
	}()
	return nil
}

// FormatSubscribeRequest creates the type for a subscribe request
func (c *Client) FormatSubscribeRequest() interface{} {
	return &subscribeRequest{
		Event: "subscribe",
		Pair:  []string{"XBT/USD", "ETH/USD"},
		Subscription: struct {
			Name string `json:"name"`
		}{Name: TICKER},
	}
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
	c.getPair(&res)
	if res.Pair != "" {
		quote = *broker.NewExchangeQuote(c.exchangeName, res.Pair, res.Price)
	}
	return quote, nil
}

func (c *Client) getPair(res *tickerResponse) {
	res.Pair = c.channelPairMap[res.ChannelID]
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

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "ws.kraken.com"}
}
