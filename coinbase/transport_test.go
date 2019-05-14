package coinbase

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/kaplanmaxe/helgart/api"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
	"github.com/kaplanmaxe/helgart/mock"
)

type mockClient struct {
	Client
}

func newMockClient(api api.Connector, quoteCh chan<- broker.Quote, errorCh chan<- error) exchange.API {
	return &mockClient{
		Client: Client{
			quoteCh:      quoteCh,
			errorCh:      errorCh,
			api:          api,
			exchangeName: exchange.COINBASE,
		},
	}
}

// Start starts the api connection and listens for new ticker messages
func (m *mockClient) Start(ctx context.Context) {
	m.Client.api.Start(ctx)
	url := m.GetURL()
	url.Scheme = "wss"
	m.Client.api.Connect(url)
	go m.StartTickerListener(ctx)
	m.Client.api.WriteMessage([]byte("echo123"))
}

func (m *mockClient) StartTickerListener(ctx context.Context) {
cLoop:
	for {
		message, err := m.Client.api.ReadMessage()
		fmt.Println(string(message))
		if err != nil {
			m.errorCh <- fmt.Errorf("Error reading from %s: %s", m.exchangeName, err)
			return
		}

		if string(message) != "echo123" {
			m.errorCh <- fmt.Errorf("Incorrect response received for gateway %s: %s", m.Client.exchangeName, err)
		} else {
			close(m.quoteCh)
		}
		break cLoop
	}
}

func (m *mockClient) GetURL() *url.URL {
	return &url.URL{Scheme: "http", Host: "127.0.0.1:457"}
}

func (m *mockClient) writeMessage(msg []byte) error {
	err := m.Client.api.WriteMessage(msg)
	if err != nil {
		return fmt.Errorf("Error writing message: %s", err)
	}
	return nil
}

func TestStart(t *testing.T) {
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()

	client := newMockClient(mock.NewMockConnector(), quoteCh, errorCh)
	client.Start(ctx)

	for {
		select {
		case <-quoteCh:
			return
		case err := <-errorCh:
			t.Fatalf("%s", err)
			return
		default:
		}
	}
}
