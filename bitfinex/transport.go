package bitfinex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	API            api.WebSocketHelper
	channelPairMap exchange.ChannelPairMap
	exchangeName   string
}

// NewClient returns a new instance of the API
func NewClient(api api.WebSocketHelper, quoteCh chan<- broker.Quote, errorCh chan<- error) *Client {
	return &Client{
		quoteCh:        quoteCh,
		errorCh:        errorCh,
		API:            api,
		channelPairMap: make(exchange.ChannelPairMap),
		exchangeName:   exchange.BITFINEX,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context) error {
	err := c.GetPairs()
	if err != nil {
		return err
	}
	c.API.Connect(c.GetURL())
	if err != nil {
		return err
	}

	doneCh := make(chan struct{}, 1)
	go c.startSubscribeListener(doneCh)
	for key, pair := range c.Pairs {
		req := &SubscriptionRequest{
			Event:   "subscribe",
			Channel: "ticker",
			Symbol:  "t" + strings.ToUpper(pair),
		}

		err := c.SendSubscribeRequest(req)
		if err != nil {
			return err
		}

		if key == len(c.Pairs)-1 {
			go func() {
				<-doneCh
				c.StartTickerListener(ctx)
			}()

		}
	}

	return nil
}

// SendSubscribeRequest overrides the interface method and sends a subscription request and listens
// for a response
func (c *Client) SendSubscribeRequest(req interface{}) error {
	// doneCh := make(chan struct{})
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Error marshalling %s subscribe request: %s", c.exchangeName, err)
	}

	err = c.API.WriteMessage(payload)
	if err != nil {
		return fmt.Errorf("Error sending subscribe request for %s: %s", c.exchangeName, err)
	}
	return nil
}

func (c *Client) startSubscribeListener(doneCh chan<- struct{}) {
	var mtx sync.Mutex
	subs := 0
Loop:
	for {
		mtx.Lock()
		message, err := c.API.ReadMessage()
		mtx.Unlock()
		if err != nil {
			c.errorCh <- fmt.Errorf("Error reading message from %s: %s", c.exchangeName, err)
			return
		}

		var subStatusResponse SubscriptionResponse
		// Skip ticker responses until we are done subscribing
		if strings.Contains(string(message), "[") {
			continue
		}

		// Weird bitfinex bug where it sends nothing?
		if len(message) == 0 {
			continue
		}

		err = json.Unmarshal(message, &subStatusResponse)
		if err != nil {
			c.errorCh <- fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
		}

		if subs < len(c.Pairs) {
			c.channelPairMap[subStatusResponse.ChannelID] = subStatusResponse.Pair
		} else {
			close(doneCh)
			break Loop
		}
		mtx.Lock()
		subs++
		mtx.Unlock()
	}
	return
}

// FormatSubscribeRequest creates the type for a subscribe request
func (c *Client) FormatSubscribeRequest() interface{} {
	return nil
}

// ParseTickerResponse parses the ticker response and returns a new instance of a broker.Quote
func (c *Client) ParseTickerResponse(msg []byte) ([]broker.Quote, error) {
	var err error
	var quotes []broker.Quote

	var res TickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		return []broker.Quote{}, fmt.Errorf("Error unmarshalling from %s: %s", c.exchangeName, err)
	}
	c.getPair(&res)
	if res.Pair != "" {
		quotes = append(quotes, *broker.NewExchangeQuote(c.exchangeName, res.Pair, res.Price))
	}
	return quotes, nil
}

func (c *Client) getPair(res *TickerResponse) {
	res.Pair = c.channelPairMap[res.ChannelID]
}

// StartTickerListener starts a new goroutine to listen for new ticker messages
func (c *Client) StartTickerListener(ctx context.Context) {
	go func() {
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
					if res[0].Price != "" {
						c.quoteCh <- res[0]
					}
				}
			}
		}
	}()
}

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "api-pub.bitfinex.com", Path: "/ws/2"}
}

// GetPairs returns all pairs for an exchange
func (c *Client) GetPairs() error {
	u := url.URL{Scheme: "https", Host: "api.bitfinex.com", Path: "/v1/symbols"}
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var pairsResponse []string
	err = json.Unmarshal(body, &pairsResponse)
	if err != nil {
		return err
	}
	var pairs []string
	for _, pair := range pairsResponse {
		pairs = append(pairs, pair)
	}
	c.Pairs = pairs
	return nil
}
