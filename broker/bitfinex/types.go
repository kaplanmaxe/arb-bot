package bitfinex

import (
	"encoding/json"
	"fmt"
)

// SubscriptionRequest represents a struct to subscribe to a ticker
// { "event": "subscribe", "channel": "ticker", "symbol": "tETHUSD" }
type SubscriptionRequest struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol"`
}

// SubscriptionResponse represents a response from the ticker subscription
// {"event":"subscribed","channel":"ticker","chanId":32034,"symbol":"tBTCUSD","pair":"BTCUSD"}
type SubscriptionResponse struct {
	ChannelID int    `json:"chanId"`
	Pair      string `json:"pair"`
}

// TickerResponse is a response from the ws api with a tick
type TickerResponse struct {
	ChannelID int
	Pair      string
	Bid       string
	Ask       string
}

// UnmarshalJSON overrides UnmarshalJSON due to Bitfinex's weird output
// [31662,[226.96,242.35969366999996,226.97,706.67194586,19.28,0.0928,226.94,290306.01961965,229.48,199.9]]
func (s *TickerResponse) UnmarshalJSON(msg []byte) error {
	// Hack for weird bitfinex response
	var channel []interface{}
	json.Unmarshal(msg, &channel)
	var tmp [][]interface{}
	json.Unmarshal(msg, &tmp)
	switch channel[1].(type) {
	case string:
		// HEARTBEAT
		return nil
	default:
		if len(tmp) != 0 {
			s.Bid = fmt.Sprintf("%f", tmp[1][0])
			s.Ask = fmt.Sprintf("%f", tmp[1][2])
			s.ChannelID = int(channel[0].(float64))
		}
	}

	return nil
}
