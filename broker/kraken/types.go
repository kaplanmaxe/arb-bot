package kraken

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Enum for subscription events
const (
	TICKER = "ticker"
	OHLC   = "ohlc"
	TRADE  = "trade"
	BOOK   = "book"
	SPREAD = "spread"
)

// SubscribeRequest represents a request to subscribe to an event
//
// Ex:
// {
// 	"event": "subscribe",
// 	"pair": [
// 	  "XBT/USD","XBT/EUR"
// 	],
// 	"subscription": {
// 	  "name": "ticker"
// 	}
//   }
type SubscribeRequest struct {
	Event        string   `json:"event"`
	Pair         []string `json:"pair"`
	Subscription struct {
		Name string `json:"name"`
	} `json:"subscription"`
}

// TickerResponse is a response from the ws api with a tick
type TickerResponse struct {
	ChannelID int
	Pair      string
	Bid       string
	Ask       string
}

// SubscriptionResponse is a response after subscribing to an event
type SubscriptionResponse struct {
	ChannelID    int    `json:"channelID"`
	Pair         string `json:"pair"`
	Status       string `json:"status"`
	Subscription struct {
		Name string `json:"name"`
	} `json:"subscription"`
}

// channelPairMap maps a channelid returned in the api to a specific pair
type channelPairMap map[int]string

type tickerResponse struct {
	Ask []interface{} `json:"a"`
	Bid []interface{} `json:"b"`
}

// UnmarshalJSON overrides UnmarshalJSON due to Kraken's weird output
func (s *TickerResponse) UnmarshalJSON(msg []byte) error {
	// Ignore heartbeats
	// TODO: can we be more efficient?
	if strings.Contains(string(msg), "heartbeat") {
		return nil
	}
	var resp []json.RawMessage
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		return fmt.Errorf("Error unmarshalling kraken TickerResponse: %s", err)
	}
	var channel int
	err = json.Unmarshal(resp[0], &channel)
	if err != nil {
		return fmt.Errorf("Error unmarshalling kraken TickerResponse: %s", err)
	}
	s.ChannelID = channel
	// Get Price
	var tickerInfo tickerResponse
	err = json.Unmarshal(resp[1], &tickerInfo)
	if err != nil {
		return fmt.Errorf("Error unmarshalling kraken TickerResponse: %s", err)
	}
	s.Ask = tickerInfo.Ask[0].(string)
	s.Bid = tickerInfo.Bid[0].(string)
	return nil
}

type assetPairResponse struct {
	Error  []string         `json:"error"`
	Result assetPairsResult `json:"result"`
}

type assetPairsResult map[string]pair

type pair struct {
	Pair          string `json:"wsname"`
	BaseCurrency  string `json:"base"`
	QuoteCurrency string `json:"quote"`
}
