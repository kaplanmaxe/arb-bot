package coinbase

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
	Subscriptions []string
	conn          *websocket.Conn
	// connResponseChannel      chan ConnectionResponse
	// subStatusResponseChannel chan SubscriptionResponse
	tickerResponseChannel chan tickerResponse
	connChannel           chan struct{}
	// channelPairMap           ChannelPairMap
}

func NewClient(pairs []string) *Client {
	return &Client{
		Subscriptions:         pairs,
		tickerResponseChannel: make(chan tickerResponse),
		connChannel:           make(chan struct{}, 1),
	}
}

// Connect connects to the websocket api and sets connection details
func (cl *Client) Connect(ctx context.Context) {
	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	connCtx, connCancel := context.WithCancel(context.Background())
	// tickerCtx, tickerCancel := context.WithCancel(context.Background())
	defer connCancel()
	// defer tickerCancel()

	cl.connect(connCtx)
	cl.subscribe()
	defer cl.conn.Close()

	// subs := 0
	for {
		select {
		// case <-cl.connResponseChannel:
		// 	connCancel()
		// 	fmt.Println("connection")
		// close(cl.connResponseChannel)
		// cl.ConnectionID = res.ConnectionID
		// cl.Version = res.Version
		// log.Printf("Connection established! Connection ID: %d, API Version: %s", cl.ConnectionID, cl.Version)
		case res := <-cl.tickerResponseChannel:

			quote := broker.NewExchangeQuote("coinbase-pro", res.Pair, res.Price)
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
