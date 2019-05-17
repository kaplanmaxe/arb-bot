package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/helgart/api"
)

// Connector is a mock connector for testing
type Connector struct {
	conn         *websocket.Conn
	exchangeName string
	URL          *url.URL
	ignoreFunc   func(msg []byte) bool
}

// NewConnector returns a new mock connector
func NewConnector(ignoreFunc func(msg []byte) bool) api.WebSocketHelper {
	return &Connector{
		exchangeName: "mockWebsocketServer",
		URL:          &url.URL{Scheme: "ws", Host: "example.com", Path: "/ws"},
		ignoreFunc:   ignoreFunc,
	}
}

// Start starts the client
func (m *Connector) Start(ctx context.Context) error {
	err := m.Connect(m.URL)
	if err != nil {
		return err
	}
	return nil
}

// Connect connects to the web socket
func (m *Connector) Connect(url *url.URL) error {
	dialer := NewWebsocketServer(m.ignoreFunc)
	c, _, err := dialer.Dial(m.URL.String(), nil)
	if err != nil {
		return fmt.Errorf("Error connection to server: %s", err)
	}
	m.conn = c

	if err != nil {
		return fmt.Errorf("Error connecting to %s: %s", m.exchangeName, err)
	}
	return nil
}

// ReadMessage reads a message from the websocket connection
func (m *Connector) ReadMessage() ([]byte, error) {
	_, message, err := m.conn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	return message, nil
}

// SendSubscribeRequest sends a subscribe request. Some exchanges require this
// and some don't
func (m *Connector) SendSubscribeRequest(req interface{}) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = m.WriteMessage(payload)
	return nil
}

// SendSubscribeRequestWithResponse sends a subscribe request and returns the response
func (m *Connector) SendSubscribeRequestWithResponse(ctx context.Context, req interface{}) ([]byte, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = m.WriteMessage(payload)
	return nil, nil
}

// WriteMessage writes a message to the websocket connection
func (m *Connector) WriteMessage(msg []byte) error {
	err := m.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}
	return nil
}

// StartTickerListener starts a listener in a new goroutine for any new quotes
// This should be overridden by each gateway
func (m *Connector) StartTickerListener(ctx context.Context) {
}

// Close closes the connection
func (m *Connector) Close() error {
	log.Printf("%s interrupt\n", m.exchangeName)
	err := m.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("%s write close: %s", m.exchangeName, err)
	}
	err = m.conn.Close()
	if err != nil {
		return fmt.Errorf("Error closing connection for %s: %s", m.exchangeName, err)
	}
	return nil
}
