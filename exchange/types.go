package exchange

import (
	"context"
	"net/url"
)

// ChannelPairMap maps a channelid returned in the api to a specific pair
type ChannelPairMap map[int]string

// Exchange represents an exchange and each exchange should implement this interface
type Exchange interface {
	Start(context.Context) error
	StartTickerListener(context.Context)
	GetURL() *url.URL
	ParseTickerResponse(msg []byte) ([]Quote, error)
}
