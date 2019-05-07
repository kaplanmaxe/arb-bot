package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/cw-websocket/kraken"
)

type Client struct {
	ConnectionID             uint64
	Version                  string
	Subscriptions            []kraken.Subscription
	responseCount            int
	mtx                      *sync.Mutex
	conn                     *websocket.Conn
	connResponseChannel      chan kraken.ConnectionResponse
	subStatusResponseChannel chan kraken.SubscriptionResponse
	tickerResponseChannel    chan kraken.TickerResponse
}

func NewClient(s []kraken.Subscription) *Client {
	return &Client{
		Subscriptions:            s,
		mtx:                      &sync.Mutex{},
		connResponseChannel:      make(chan kraken.ConnectionResponse, 1),
		subStatusResponseChannel: make(chan kraken.SubscriptionResponse, 1),
		tickerResponseChannel:    make(chan kraken.TickerResponse),
	}
}

func (cl *Client) Connect() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	connCtx, connCancel := context.WithCancel(context.Background())
	subStatusCtx, subStatusCancel := context.WithCancel(context.Background())
	tickerCtx, tickerCancel := context.WithCancel(context.Background())
	defer connCancel()
	defer subStatusCancel()
	defer tickerCancel()

	cl.connect(connCtx)
	defer cl.conn.Close()

	subs := 0
	for {
		select {
		case res := <-cl.connResponseChannel:
			connCancel()
			// close(cl.connResponseChannel)
			cl.ConnectionID = res.ConnectionID
			cl.Version = res.Version
			log.Printf("Connection established! Connection ID: %d, API Version: %s", cl.ConnectionID, cl.Version)
			cl.subscribe(subStatusCtx)
		case res := <-cl.subStatusResponseChannel:
			subs++
			if subs == len(cl.Subscriptions) {
				subStatusCancel()
				fmt.Println("ok")
				cl.startTickerListener(tickerCtx)
			}
			// Minor race condition where we can't stop go routine fast enough
			if res.Pair != "" {
				log.Printf("%s subscribed to for pair %s on channel %d", res.Subscription.Name, res.Pair, res.ChannelID)
			}
		case res := <-cl.tickerResponseChannel:
			log.Printf("Test: %#v", res)
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
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
	u := url.URL{Scheme: "wss", Host: "ws.kraken.com"}
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

			var connResponse kraken.ConnectionResponse
			err = json.Unmarshal(message, &connResponse)
			if err != nil {
				// TODO: fix
				log.Fatal("Unmarshall err2", err)
			}

			cl.connResponseChannel <- connResponse

			select {
			case <-ctx.Done():
				break cLoop
			}
		}
	}()
}

func (cl *Client) subscribe(ctx context.Context) {
	// TODO: subscribe all at once
	for _, sub := range cl.Subscriptions {
		req := &kraken.SubscribeRequest{
			Event:        "subscribe",
			Pair:         sub.Pair,
			Subscription: kraken.SubscriptionT{Name: sub.Type},
		}
		payload, err := json.Marshal(req)
		if err != nil {
			log.Fatal("Marshal", err)
		}
		err = cl.conn.WriteMessage(websocket.TextMessage, []byte(payload))
		if err != nil {
			log.Fatal("Write", err)
		}
	}

	go func() {
	Loop:
		for {
			message, err := cl.readMessage()
			if err != nil {
				// TODO: fix
				log.Println("read1:", err, message)
				return
			}

			var subStatusResponse kraken.SubscriptionResponse
			err = json.Unmarshal(message, &subStatusResponse)
			if err != nil {
				// TODO: fix
				log.Fatal("Unmarshall err1", err)
			}
			cl.subStatusResponseChannel <- subStatusResponse

			select {
			case <-ctx.Done():
				break Loop
			default:
			}
		}
	}()
}

func (cl *Client) startTickerListener(ctx context.Context) {
	go func() {
	Loop:
		for {
			message, err := cl.readMessage()
			if err != nil {
				// TODO: fix
				log.Println("read error, skipping")
			}
			if len(message) == 0 {
				fmt.Println("skip")
			}
			var tickerResponse kraken.TickerResponse
			err = json.Unmarshal(message, &tickerResponse)
			if err != nil {
				// TODO: fix
				log.Fatal("Unmarshall err3", err, message)
			}
			if tickerResponse.Ask != "" {
				cl.tickerResponseChannel <- tickerResponse
			}

			select {
			case <-ctx.Done():
				break Loop
			default:
			}
		}
	}()
}
