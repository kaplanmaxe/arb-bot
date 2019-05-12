package coinbase

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

// NewClient returns a new instance of a coinbase api client
func NewClient(pairs []string) *Client {
	return &Client{
		Subscriptions: pairs,
	}
}

// Connect connects to the websocket api and sets connection details
func (cl *Client) Connect(ctx context.Context, quoteCh chan<- broker.Quote) {
	cl.connect(ctx, quoteCh)
	cl.subscribe()
}

func (cl *Client) readMessage() ([]byte, error) {
	_, message, err := cl.conn.ReadMessage()
	if err != nil {
		return []byte{}, err
	}
	return message, nil
}

func (cl *Client) connect(ctx context.Context, quoteCh chan<- broker.Quote) {
	u := url.URL{Scheme: "wss", Host: "ws-feed.pro.coinbase.com"}
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
				log.Println("cb read2:", err, message)
				return
			}

			select {
			case <-ctx.Done():
				log.Println("Coinbase interrupt")
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
					quoteCh <- *broker.NewExchangeQuote("coinbase-pro", res.Pair, res.Price)
				}
			}
		}
	}()
}

func (cl *Client) subscribe() error {
	subRequest := &subscribeRequest{
		Type:       "subscribe",
		ProductIDs: cl.Subscriptions,
		Channels: []struct {
			Name       string   `json:"name"`
			ProductIDs []string `json:"product_ids"`
		}{
			{
				Name:       "ticker",
				ProductIDs: cl.Subscriptions,
			},
		},
	}
	payload, err := json.Marshal(subRequest)
	if err != nil {
		return err
	}
	err = cl.conn.WriteMessage(websocket.TextMessage, []byte(payload))
	if err != nil {
		return err
	}
	return nil
}
