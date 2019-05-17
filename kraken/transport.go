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
	"github.com/kaplanmaxe/helgart/exchange"
)

// Client represents an API client
type Client struct {
	Pairs          []string
	quoteCh        chan<- exchange.Quote
	errorCh        chan<- error
	API            api.WebSocketHelper
	channelPairMap exchange.ChannelPairMap
	exchangeName   string
	productMap     exchange.ExchangeProductMap
}

// NewClient returns a new instance of the API
func NewClient(api api.WebSocketHelper, quoteCh chan<- exchange.Quote, errorCh chan<- error) *Client {
	return &Client{
		quoteCh:        quoteCh,
		errorCh:        errorCh,
		API:            api,
		channelPairMap: make(exchange.ChannelPairMap),
		exchangeName:   exchange.KRAKEN,
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context, productMap exchange.ProductMap) error {
	c.productMap = productMap[c.exchangeName]
	err := c.GetPairs()
	if err != nil {
		return err
	}
	c.API.Connect(c.GetURL())
	if err != nil {
		return err
	}

	doneCh, err := c.SendSubscribeRequest(c.FormatSubscribeRequest())
	if err != nil {
		return err
	}
	go func() {
		<-doneCh
		c.StartTickerListener(ctx)
	}()

	return nil
}

// SendSubscribeRequest overrides the interface method and sends a subscription request and listens
// for a response
func (c *Client) SendSubscribeRequest(req interface{}) (<-chan struct{}, error) {
	doneCh := make(chan struct{})
	payload, err := json.Marshal(req)
	if err != nil {
		return doneCh, fmt.Errorf("Error marshalling %s subscribe request: %s", c.exchangeName, err)
	}
	err = c.API.WriteMessage(payload)
	if err != nil {
		return doneCh, fmt.Errorf("Error sending subscribe request for %s: %s", c.exchangeName, err)
	}
	go func() {
		var mtx sync.Mutex
		subs := 0
	Loop:
		for {
			message, err := c.API.ReadMessage()
			if err != nil {
				c.errorCh <- fmt.Errorf("Error reading message from %s: %s", c.exchangeName, err)
				return
			}

			var subStatusResponse SubscriptionResponse

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
	}()
	return doneCh, nil
}

// FormatSubscribeRequest creates the type for a subscribe request
func (c *Client) FormatSubscribeRequest() interface{} {
	return &SubscribeRequest{
		Event: "subscribe",
		Pair:  c.Pairs,
		Subscription: struct {
			Name string `json:"name"`
		}{Name: TICKER},
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
	c.getPair(&res)
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
					if res[0].HePair != "" {
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
	c.Pairs = pairs
	return nil
}
