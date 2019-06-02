package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	orderBookMap exchange.OrderBookMap
}

// NewClient returns a new instance of the API
func NewClient(api api.WebSocketHelper, quoteCh chan<- exchange.Quote, errorCh chan<- error) *Client {
	return &Client{
		quoteCh:      quoteCh,
		errorCh:      errorCh,
		API:          api,
		exchangeName: exchange.COINBASE,
		orderBookMap: make(exchange.OrderBookMap),
	}
}

// Start starts the api connection and listens for new ticker messages
func (c *Client) Start(ctx context.Context, productMap exchange.ProductMap, doneCh chan<- struct{}) error {
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
	go c.StartTickerListener(ctx, doneCh)
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
				Name:       "level2",
				ProductIDs: c.pairs,
			},
		},
	}
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
				if res[0].HePair != "" {
					c.quoteCh <- res[0]
				}
			}
		}
	}
}

// ParseTickerResponse parses the ticker response and returns a new instance of a exchange.Quote
func (c *Client) ParseTickerResponse(message []byte) ([]exchange.Quote, error) {
	var quotes []exchange.Quote
	// On snapshot response we store best bid and ask per pair
	if strings.Contains(string(message), "snapshot") {
		var snapshotResponse SnapshotResponse
		err := json.Unmarshal(message, &snapshotResponse)
		if err != nil {
			return []exchange.Quote{}, fmt.Errorf("Error unmarshalling snapshot response for %s", c.exchangeName)
		}
		// Initialize bids and asks stacks
		if _, ok := c.orderBookMap[snapshotResponse.Pair]; !ok {
			pair := c.orderBookMap[snapshotResponse.Pair]
			pair.Bids = exchange.NewSpreadStack(40, "bid")
			pair.Asks = exchange.NewSpreadStack(40, "ask")
			c.orderBookMap[snapshotResponse.Pair] = pair
		}
		// Initialize bid side
		var length int
		if len(snapshotResponse.Bids) >= 20 {
			length = 20
		} else {
			length = len(snapshotResponse.Bids)
		}
		for i := 0; i < length; i++ {
			price, err := strconv.ParseFloat(snapshotResponse.Bids[i][0], 64)
			if err != nil {
				return []exchange.Quote{}, fmt.Errorf("Error parsing order book values from %s", c.exchangeName)
			}
			size, err := strconv.ParseFloat(snapshotResponse.Bids[i][1], 64)
			if err != nil {
				return []exchange.Quote{}, fmt.Errorf("Error parsing order book values from %s", c.exchangeName)
			}
			c.orderBookMap[snapshotResponse.Pair].Bids.Push(&exchange.SpreadNode{
				Price: price,
				Size:  size,
			})
		}
		if len(snapshotResponse.Asks) >= 20 {
			length = 20
		} else {
			length = len(snapshotResponse.Asks)
		}
		for i := 0; i < length; i++ {
			price, err := strconv.ParseFloat(snapshotResponse.Asks[i][0], 64)
			if err != nil {
				return []exchange.Quote{}, fmt.Errorf("Error parsing order book values from %s", c.exchangeName)
			}
			size, err := strconv.ParseFloat(snapshotResponse.Asks[i][1], 64)
			if err != nil {
				return []exchange.Quote{}, fmt.Errorf("Error parsing order book values from %s", c.exchangeName)
			}
			c.orderBookMap[snapshotResponse.Pair].Asks.Push(&exchange.SpreadNode{
				Price: price,
				Size:  size,
			})
		}
	} else if strings.Contains(string(message), "l2update") {
		var response LevelTwoResponse
		err := json.Unmarshal(message, &response)
		if err != nil {
			return []exchange.Quote{}, fmt.Errorf("Error unmarshalling level two response for %s", c.exchangeName)
		}
		// If no pair name skip everything and return
		if response.Pair == "" {
			return []exchange.Quote{}, nil
		}
		for _, val := range response.Changes {
			price, err := strconv.ParseFloat(val[1], 64)
			if err != nil {
				return []exchange.Quote{}, fmt.Errorf("Error reading new quote from %s", c.exchangeName)
			}
			size, err := strconv.ParseFloat(val[2], 64)
			if err != nil {
				return []exchange.Quote{}, fmt.Errorf("Error reading new quote from %s", c.exchangeName)
			}
			cachedBestBid := c.orderBookMap[response.Pair].Bids.Nodes[0].Price
			cachedBestAsk := c.orderBookMap[response.Pair].Asks.Nodes[0].Price
			buySide := c.orderBookMap[response.Pair].Bids
			sellSide := c.orderBookMap[response.Pair].Asks
			if val[0] == "buy" {
				if size == 0 {
					// If size is 0 we remove it
					buySide.Pop(price)
				} else if price > buySide.Nodes[len(buySide.Nodes)-1].Price && size > 0 {
					// If on the buy side, the price is greater than the last node, and the size isn't 0, we add
					buySide.Push(&exchange.SpreadNode{
						Price: price,
						Size:  size,
					})
				}
			} else if val[0] == "sell" {
				if size == 0 {
					// If size is 0 we remove it
					sellSide.Pop(price)
				} else if price < sellSide.Nodes[len(sellSide.Nodes)-1].Price && size > 0 {
					// If on the buy side, the price is greater than the last node, and the size isn't 0, we add
					sellSide.Push(&exchange.SpreadNode{
						Price: price,
						Size:  size,
					})
				}
			}
			// TODO: how does this get empty?
			if len(c.orderBookMap[response.Pair].Bids.Nodes) == 0 || len(c.orderBookMap[response.Pair].Asks.Nodes) == 0 {
				log.Printf("Orderbook on %s is empty for %s\n", c.exchangeName, response.Pair)
				return []exchange.Quote{}, nil
			}
			if c.orderBookMap[response.Pair].Bids.Nodes[0].Price > cachedBestBid ||
				c.orderBookMap[response.Pair].Asks.Nodes[0].Price < cachedBestAsk {

				product := c.productMap[response.Pair]
				quotes = append(quotes, exchange.Quote{
					Exchange: c.exchangeName,
					// TODO: remove the string conversion
					Bid:     fmt.Sprintf("%8.8f", c.orderBookMap[response.Pair].Bids.Nodes[0].Price),
					Ask:     fmt.Sprintf("%8.8f", c.orderBookMap[response.Pair].Asks.Nodes[0].Price),
					ExPair:  product.ExPair,
					HePair:  product.HePair,
					ExBase:  product.ExBase,
					HeBase:  product.HeBase,
					ExQuote: product.ExQuote,
					HeQuote: product.HeQuote,
				})
			}
		}
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
