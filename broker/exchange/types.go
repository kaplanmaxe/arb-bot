package exchange

import (
	"context"
	"net/url"
)

// ChannelPairMap maps a channelid returned in the api to a specific pair
type ChannelPairMap map[int]string

// Exchange represents an exchange and each exchange should implement this interface
type Exchange interface {
	Start(context.Context, ProductMap, chan<- struct{}) error
	StartTickerListener(context.Context, chan<- struct{})
	GetURL() *url.URL
	ParseTickerResponse(msg []byte) ([]Quote, error)
}
