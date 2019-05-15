package kraken

import (
	"encoding/json"
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

type TickerResponse struct {
	ChannelID int
	Pair      string
	Price     string
}

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

func (s *TickerResponse) UnmarshalJSON(msg []byte) error {
	// Hack for weird kraken response
	var channel []int
	json.Unmarshal(msg, &channel)
	var tmp []map[string][]interface{}
	json.Unmarshal(msg, &tmp)
	if len(tmp) != 0 {
		s.Price = tmp[1]["a"][0].(string)
		s.ChannelID = channel[0]
	}

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
