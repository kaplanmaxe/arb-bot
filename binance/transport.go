package binance

import (
	"context"
	"encoding/json"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/cw-websocket/broker"
)

// Client represents a new websocket client for Coinbase
type Client struct {
	Subscriptions []string
	conn          *websocket.Conn
}

// NewClient returns a new instance of a binance api client
func NewClient(pairs []string) *Client {
	return &Client{
		Subscriptions: pairs,
	}
}

// Connect connects to the websocket api and sets connection details
func (cl *Client) Connect(ctx context.Context, quoteCh chan<- broker.Quote) {
	cl.connect(ctx, quoteCh)
}

func (cl *Client) readMessage() ([]byte, error) {
	_, message, err := cl.conn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	return message, nil
}

func (cl *Client) connect(ctx context.Context, quoteCh chan<- broker.Quote) {
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

			select {
			case <-ctx.Done():
				log.Println("Binance interrupt")
				err := cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				cl.conn.Close()
				break cLoop
			default:
				var res tickerResponse
				err = json.Unmarshal(message, &res)
				if err != nil {
					log.Fatal("Unmarshal", err)
				}
				if res.Pair != "" {
					quoteCh <- *broker.NewExchangeQuote("binance", res.Pair, res.Price)
				}
			}
		}
	}()
}
