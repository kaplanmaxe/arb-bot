package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

// Client represents an API client
type Client struct {
	pairs          []string
	quoteCh        chan<- broker.Quote
	errorCh        chan<- error
	api            api.Connector
	channelPairMap channelPairMap
	exchangeName   string
}

// NewClient returns a new instance of the API
func NewClient(api api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) exchange.API {
	return &Client{
		quoteCh:        quoteCh,
		errorCh:        errorCh,
		api:            api,
		channelPairMap: make(channelPairMap),
		exchangeName:   exchange.KRAKEN,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context) error {
	err := c.GetPairs()
	if err != nil {
		return err
	}
	c.api.Connect(c.GetURL())
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	err = c.SendSubscribeRequest(&wg, c.FormatSubscribeRequest())
	if err != nil {
		return err
	}
	wg.Wait()
	c.StartTickerListener(ctx)
	return nil
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
			if subs < len(c.pairs) {
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
		Pair:  c.pairs,
		Subscription: struct {
			Name string `json:"name"`
		}{Name: TICKER},
	}
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
	c.getPair(&res)
	if res.Pair != "" {
		quotes = append(quotes, *broker.NewExchangeQuote(c.exchangeName, res.Pair, res.Price))
	}
	return quotes, nil
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
				} else if len(res) > 0 {
					if res[0].Pair != "" {
						c.quoteCh <- res[0]
					}
				}
			}
		}
	}()
}

// GetURL returns the url for the websocket connection
func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "ws.kraken.com"}
}

// GetPairs returns all pairs for an exchange
func (c *Client) GetPairs() error {
	u := url.URL{Scheme: "https", Host: "api.kraken.com", Path: "/0/public/AssetPairs"}
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var pairsResponse assetPairResponse
	err = json.Unmarshal(body, &pairsResponse)
	if err != nil {
		return err
	}
	var pairs []string
	for key := range pairsResponse.Result {
		pairs = append(pairs, pairsResponse.Result[key].Pair)
	}
	c.pairs = pairs
	return nil
}
