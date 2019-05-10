package kraken

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/cw-websocket/broker"
	// "github.com/kaplanmaxe/cw-websocket/kraken"
)

// Client represents a new websocket client for Kraken
type Client struct {
	ConnectionID             uint64
	Version                  string
	Subscriptions            []Subscription
	conn                     *websocket.Conn
	connResponseChannel      chan ConnectionResponse
	subStatusResponseChannel chan SubscriptionResponse
	tickerResponseChannel    chan TickerResponse
	channelPairMap           ChannelPairMap
}

// NewClient returns a new instance of Client
func NewClient(s []Subscription) *Client {
	return &Client{
		Subscriptions:            s,
		connResponseChannel:      make(chan ConnectionResponse, 1),
		subStatusResponseChannel: make(chan SubscriptionResponse, 1),
		tickerResponseChannel:    make(chan TickerResponse),
		channelPairMap:           make(ChannelPairMap),
	}
}

// Connect connects to the websocket api and sets connection details
func (cl *Client) Connect(ctx context.Context) {
	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	connCtx, connCancel := context.WithCancel(context.Background())
	subStatusCtx, subStatusCancel := context.WithCancel(context.Background())
	tickerCtx, tickerCancel := context.WithCancel(context.Background())
	defer connCancel()
	defer subStatusCancel()
	defer tickerCancel()

	cl.connect(connCtx)

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
				cl.startTickerListener(tickerCtx)
			}
			// Minor race condition where we can't stop go routine fast enough
			if res.Pair != "" {
				cl.channelPairMap[res.ChannelID] = res.Pair
				log.Printf("%s subscribed to for pair %s on channel %d", res.Subscription.Name, res.Pair, res.ChannelID)
			}
		case res := <-cl.tickerResponseChannel:
			quote := broker.NewExchangeQuote("kraken", cl.channelPairMap[res.ChannelID], res.Ask)
			log.Printf("Quote: %#v", quote)
		// case <-interrupt:
		case <-ctx.Done():
			log.Println("interrupt")
			defer connCancel()
			defer subStatusCancel()
			defer tickerCancel()
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

			var connResponse ConnectionResponse
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
		req := &SubscribeRequest{
			Event:        "subscribe",
			Pair:         sub.Pair,
			Subscription: SubscriptionT{Name: sub.Type},
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

			var subStatusResponse SubscriptionResponse
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
				return
			}
			var tickerResponse TickerResponse
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
