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

// SpreadResponse is a response from the ws api with a tick
type SpreadResponse struct {
	Pair      string
	Bid       string
	Ask       string
	BidVolume string
	AskVolume string
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

// UnmarshalJSON overrides UnmarshalJSON due to Kraken's weird output
func (s *SpreadResponse) UnmarshalJSON(msg []byte) error {
	// Ignore heartbeats, initial connection response, and subscription response
	// TODO: can we be more efficient?
	if strings.Contains(string(msg), "heartbeat") || strings.Contains(string(msg), "connectionID") ||
		strings.Contains(string(msg), "channelID") {
		return nil
	}
	var resp []json.RawMessage
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		return fmt.Errorf("Error unmarshalling kraken SpreadResponse: %s", err)
	}
	var pair string
	err = json.Unmarshal(resp[3], &pair)
	if err != nil {
		return fmt.Errorf("Error unmarshalling kraken SpreadResponse pair: %s", err)
	}
	// s.ChannelID = channel
	// Get best bid and ask
	var spread []string
	err = json.Unmarshal(resp[1], &spread)
	if err != nil {
		return fmt.Errorf("Error unmarshalling kraken SpreadResponse spread: %s", err)
	}
	s.Pair = pair
	s.Bid = spread[0]
	s.Ask = spread[1]
	s.BidVolume = spread[3]
	s.AskVolume = spread[4]
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
