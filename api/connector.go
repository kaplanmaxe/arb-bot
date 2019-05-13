package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type Connector interface {
	Start(context.Context)
	Connect(*url.URL) error
	ReadMessage() ([]byte, error)
	SendSubscribeRequest(interface{}) error
	SendSubscribeRequestWithResponse(context.Context, interface{}) ([]byte, error)
	WriteMessage([]byte) error
	StartTickerListener(context.Context)
	Close() error
}

// Source represents an exchange source
type Source struct {
	Pairs        []string
	conn         *websocket.Conn
	exchangeName string
	// quoteCh      chan<- broker.Quote
	// api          exchange.API
}

// NewSource returns a new instance of source
func NewSource(exchangeName string) Connector {
	return &Source{
		exchangeName: exchangeName,
		// quoteCh:      quoteCh,
		// api:          api,
	}
}

func (s *Source) Start(ctx context.Context) {}

// Connect connects to the websocket api and stores the connection
func (s *Source) Connect(url *url.URL) error {
	c, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	s.conn = c

	if err != nil {
		return fmt.Errorf("Error connecting to %s: %s", s.exchangeName, err)
	}
	return nil
}

func (s *Source) ReadMessage() ([]byte, error) {
	_, message, err := s.conn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	return message, nil
}

// SendSubscribeRequest sends a subscribe request. Some exchanges require this
// and some don't
func (s *Source) SendSubscribeRequest(req interface{}) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	err = s.WriteMessage(payload)
	return nil
}

// SendSubscribeRequestWithResponse sends a subscribe request and returns the response
func (s *Source) SendSubscribeRequestWithResponse(ctx context.Context, req interface{}) ([]byte, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = s.WriteMessage(payload)
	return nil, nil
}

func (s *Source) WriteMessage(msg []byte) error {
	err := s.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}
	return nil
}

// StartTickerListener starts a listener in a new goroutine for any new quotes
// This should be overridden by each gateway
func (s *Source) StartTickerListener(ctx context.Context) {
}

// Close closes the connection
func (s *Source) Close() error {
	log.Printf("%s interrupt\n", s.exchangeName)
	err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("%s write close: %s", s.exchangeName, err)
	}
	err = s.conn.Close()
	if err != nil {
		return fmt.Errorf("Error closing connection for %s: %s", s.exchangeName, err)
	}
	return nil
}
