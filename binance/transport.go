package binance

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/cw-websocket/broker"
)

// Client represents a new websocket client for Coinbase
type Client struct {
	Subscriptions         []string
	conn                  *websocket.Conn
	tickerResponseChannel chan tickerResponse
}

func NewClient(pairs []string) *Client {
	return &Client{
		Subscriptions:         pairs,
		tickerResponseChannel: make(chan tickerResponse),
	}
}

// Connect connects to the websocket api and sets connection details
func (cl *Client) Connect(ctx context.Context) {
	connCtx, connCancel := context.WithCancel(context.Background())
	defer connCancel()

	cl.connect(connCtx)
	defer cl.conn.Close()

	// subs := 0
	for {
		select {
		case res := <-cl.tickerResponseChannel:

			quote := broker.NewExchangeQuote("binance", res.Pair, res.Price)
			log.Printf("Quote: %#v", quote)
		// case <-interrupt:
		case <-ctx.Done():
			log.Println("interrupt")
			// defer connCancel()
			// defer subStatusCancel()
			// defer tickerCancel()
			defer cl.conn.Close()
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (cl *Client) readMessage() ([]byte, error) {
	_, message, err := cl.conn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	return message, nil
}

func (cl *Client) connect(ctx context.Context) {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws/bnbbtc@ticker"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	cl.conn = c

	if err != nil {
		log.Fatal("dial:", err)
	}
	go func() {
	cLoop:
		for {
			message, err := cl.readMessage()
			if err != nil {
				// TODO: fix
				log.Println("read2:", err, message)
				return
			}

			var res tickerResponse
			err = json.Unmarshal(message, &res)
			if err != nil {
				log.Fatal("Unmarshal", err)
			}
			if res.Pair != "" {
				cl.tickerResponseChannel <- res
			}

			select {
			case <-ctx.Done():
				break cLoop
			default:
			}
		}
	}()
}
