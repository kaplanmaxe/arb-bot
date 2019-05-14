package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

type mockClient struct {
	quoteCh      chan<- broker.Quote
	errorCh      chan<- error
	Client       api.Connector
	ExchangeName string
	exchange.API
}

func newMockClient(a api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) *mockClient {
	return &mockClient{
		quoteCh:      quoteCh,
		errorCh:      errorCh,
		ExchangeName: "MOCK",
		Client:       a,
	}
}

// Start starts the api connection and listens for new ticker messages
func (m *mockClient) Start(ctx context.Context) {
	url := m.GetURL()
	url.Scheme = "wss"
	err := m.Client.Connect(url)
	if err != nil {
		log.Fatal(err)
	}
	go m.StartTickerListener(ctx)
}

func (m *mockClient) StartTickerListener(ctx context.Context) {
cLoop:
	for {
		msg, err := m.Client.ReadMessage()
		if err != nil {
			m.errorCh <- fmt.Errorf("Error reading from %s: %s", m.ExchangeName, err)
			return
		}

		var quotes []broker.Quote
		err = json.Unmarshal(msg, &quotes)
		if err != nil {
			m.errorCh <- fmt.Errorf("Incorrect response received for gateway %s: %s", m.ExchangeName, err)
		} else {
			m.quoteCh <- quotes[0]
		}
		break cLoop
	}
}

func (m *mockClient) GetURL() *url.URL {
	return &url.URL{Scheme: "http", Host: "127.0.0.1:457"}
}

func (m *mockClient) writeMessage(msg []byte) error {
	err := m.Client.WriteMessage(msg)
	if err != nil {
		return fmt.Errorf("Error writing message: %s", err)
	}
	return nil
}
