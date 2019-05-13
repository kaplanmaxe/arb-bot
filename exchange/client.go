package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/helgart/broker"
)

const (
	// KRAKEN represents the kraken api
	KRAKEN = "kraken"
	// COINBASE represents the coinbase api
	COINBASE = "coinbase"
	// BINANCE represents the binance api
	BINANCE = "binance"
	// BITFINEX represents the bitfinex api
	BITFINEX = "bitfinex"
)

// API is an interface each exchange client should satisfy
type API interface {
	// Start(context.Context)
	GetURL() *url.URL
	ParseTickerResponse([]byte) broker.Quote
	FormatSubscribeRequest() interface{}
	// connect()
	// Close() error
}

// Connector is an interface containing methods to perform various actions
// on ticker websocket connections
type Connector interface {
	Start(context.Context)
	Connect(context.Context) error
	readMessage() ([]byte, error)
	SendSubscribeRequest(interface{}) error
	writeMessage([]byte) error
	StartTickerListener(context.Context)
	Close() error
}

// Source represents an exchange source
type Source struct {
	// url          url.URL
	pairs        []string
	conn         *websocket.Conn
	exchangeName string
	quoteCh      chan<- broker.Quote
	api          API
}

// NewSource returns a new instance of source
func NewSource(api API, exchangeName string, quoteCh chan<- broker.Quote) Connector {
	return &Source{
		// url:          getURL(exchangeName),
		exchangeName: exchangeName,
		quoteCh:      quoteCh,
		api:          api,
	}
}

// Start starts a new api connection
func (s *Source) Start(ctx context.Context) {
	switch s.exchangeName {
	case COINBASE:
		s.Connect(ctx)
		s.SendSubscribeRequest(s.api.FormatSubscribeRequest())
		s.StartTickerListener(ctx)
	}
}

// Connect connects to the websocket api and stores the connection
func (s *Source) Connect(ctx context.Context) error {
	c, _, err := websocket.DefaultDialer.Dial(s.api.GetURL().String(), nil)
	s.conn = c

	if err != nil {
		return fmt.Errorf("Error connecting to %s: %s", s.exchangeName, err)
	}
	return nil
}

func (s *Source) readMessage() ([]byte, error) {
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
	err = s.writeMessage(payload)
	return nil
}

func (s *Source) writeMessage(msg []byte) error {
	err := s.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}
	return nil
}

// StartTickerListener starts a listener in a new goroutine for any new quotes
func (s *Source) StartTickerListener(ctx context.Context) {
	go func() {
	cLoop:
		for {
			message, err := s.readMessage()
			if err != nil {
				// TODO: fix
				log.Println("cb read2:", err, message)
				return
			}

			select {
			case <-ctx.Done():
				err := s.Close()
				if err != nil {
					log.Printf("Error closing %s: %s", s.exchangeName, err)
				}
				break cLoop
			default:
				s.quoteCh <- s.api.ParseTickerResponse(message)
			}
		}
	}()
}

// Close closes the connection
func (s *Source) Close() error {
	log.Printf("%s interrupt\n", s.exchangeName)
	err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		// log.Printf("%s write close: %s", s.exchangeName, err)
		return fmt.Errorf("%s write close: %s", s.exchangeName, err)
	}
	s.conn.Close()
	return nil
}
