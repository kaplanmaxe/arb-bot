package kraken

import (
	"encoding/json"
)

// GeneralRequest is a struct representing a basic message request
// Such as a ping
type GeneralRequest struct {
	Event string `json:"event"`
	ReqID int    `json:"reqid"`
}

// Enum for subscription events
const (
	TICKER = "ticker"
	OHLC   = "ohlc"
	TRADE  = "trade"
	BOOK   = "book"
	SPREAD = "spread"
)

// SubscriptionT is a struct to be used in SubscribeRequest
// TODO: rename
type SubscriptionT struct {
	Name string `json:"name"`
}

type Subscription struct {
	Type string
	Pair string
}

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
	Event        string        `json:"event"`
	Pair         []string      `json:"pair"`
	Subscription SubscriptionT `json:"subscription"`
}

// TickerResponse is a struct representing a response from the ticker event
type TickerResponse struct {
	Ask string
	Bid string
}

type SubscriptionResponse struct {
	ChannelID    int    `json:"channelID"`
	Pair         string `json:"pair"`
	Status       string `json:"status"`
	Subscription struct {
		Name string `json:"name"`
	} `json:"subscription"`
}

type ConnectionResponse struct {
	ConnectionID uint64 `json:"connectionID"`
	Version      string `json:"version"`
}

func (s *TickerResponse) UnmarshalJSON(msg []byte) error {
	var tmp []map[string][]interface{}
	err := json.Unmarshal(msg, &tmp)
	if len(tmp) != 0 {
		s.Ask = tmp[1]["a"][0].(string)
		s.Bid = tmp[1]["b"][0].(string)
	}

	return err
}
