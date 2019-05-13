package kraken

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/exchange"
)

type Client struct {
	Pairs []string
}

func NewClient(pairs []string) exchange.API {
	return &Client{
		Pairs: pairs,
	}
}

func (c *Client) FormatSubscribeRequest() interface{} {
	return &subscribeRequest{
		Event: "subscribe",
		Pair:  []string{"XBT/USD", "ETH/USD"},
		Subscription: struct {
			Name string `json:"name"`
		}{Name: TICKER},
	}
}

func (c *Client) ParseTickerResponse(msg []byte) broker.Quote {
	var err error
	var quote broker.Quote

	var res tickerResponse
	err = json.Unmarshal(msg, &res)
	if err != nil {
		log.Fatal("Unmarshal", err)
	}
	if res.Pair != "" {
		quote = *broker.NewExchangeQuote("kraken", res.Pair, res.Price)
	}
	return quote
}

func (c *Client) GetURL() *url.URL {
	return &url.URL{Scheme: "wss", Host: "ws.kraken.com"}
}

// // Client represents a new websocket client for Kraken
// type Client struct {
// 	ConnectionID             uint64
// 	Version                  string
// 	Subscriptions            []Subscription
// 	conn                     *websocket.Conn
// 	connResponseChannel      chan ConnectionResponse
// 	subStatusResponseChannel chan SubscriptionResponse
// 	channelPairMap           ChannelPairMap
// }

// // NewClient returns a new instance of Client
// func NewClient(s []Subscription) *Client {
// 	return &Client{
// 		Subscriptions:            s,
// 		connResponseChannel:      make(chan ConnectionResponse, 1),
// 		subStatusResponseChannel: make(chan SubscriptionResponse, 1),
// 		channelPairMap:           make(ChannelPairMap),
// 	}
// }

// // Connect connects to the websocket api and sets connection details
// func (cl *Client) Connect(ctx context.Context, quoteCh chan<- broker.Quote) {
// 	connCtx, connCancel := context.WithCancel(context.Background())
// 	subStatusCtx, subStatusCancel := context.WithCancel(context.Background())
// 	defer connCancel()
// 	defer subStatusCancel()

// 	cl.connect(connCtx)

// 	subs := 0
// Loop:
// 	for {
// 		select {
// 		case res := <-cl.connResponseChannel:
// 			connCancel()
// 			cl.ConnectionID = res.ConnectionID
// 			cl.Version = res.Version
// 			log.Printf("Kraken connection established! Connection ID: %d, API Version: %s", cl.ConnectionID, cl.Version)
// 			cl.subscribe(subStatusCtx)
// 		case res := <-cl.subStatusResponseChannel:
// 			subs++
// 			if res.Pair != "" {
// 				cl.channelPairMap[res.ChannelID] = res.Pair
// 				log.Printf("%s subscribed to for pair %s on channel %d", res.Subscription.Name, res.Pair, res.ChannelID)
// 				// TODO: fix the hacky 0 index
// 				if subs == len(cl.Subscriptions[0].Pair) {
// 					subStatusCancel()
// 					cl.startTickerListener(ctx, quoteCh)
// 					break Loop
// 				}
// 			}
// 		}
// 	}
// }

// func (cl *Client) readMessage() ([]byte, error) {
// 	_, message, err := cl.conn.ReadMessage()
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	return message, nil
// }

// func (cl *Client) connect(ctx context.Context) {
// 	u := url.URL{Scheme: "wss", Host: "ws.kraken.com"}
// 	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	cl.conn = c
// 	if err != nil {
// 		log.Fatal("dial:", err)
// 	}
// 	go func() {
// 	cLoop:
// 		for {
// 			message, err := cl.readMessage()
// 			if err != nil {
// 				// TODO: fix
// 				log.Println("read2:", err, message)
// 				return
// 			}

// 			var connResponse ConnectionResponse
// 			err = json.Unmarshal(message, &connResponse)
// 			if err != nil {
// 				// TODO: fix
// 				log.Fatal("Unmarshall err2", err)
// 			}

// 			cl.connResponseChannel <- connResponse

// 			select {
// 			case <-ctx.Done():
// 				break cLoop
// 			}
// 		}
// 		return
// 	}()
// }

// func (cl *Client) subscribe(ctx context.Context) {
// 	// TODO: subscribe all at once
// 	for _, sub := range cl.Subscriptions {
// 		req := &SubscribeRequest{
// 			Event:        "subscribe",
// 			Pair:         sub.Pair,
// 			Subscription: SubscriptionT{Name: sub.Type},
// 		}
// 		payload, err := json.Marshal(req)
// 		if err != nil {
// 			log.Fatal("Marshal", err)
// 		}
// 		err = cl.conn.WriteMessage(websocket.TextMessage, []byte(payload))
// 		if err != nil {
// 			log.Fatal("Write", err)
// 		}
// 	}

// 	go func() {
// 	Loop:
// 		for {
// 			message, err := cl.readMessage()
// 			if err != nil {
// 				// TODO: fix
// 				log.Println("read1:", err, message)
// 				return
// 			}

// 			var subStatusResponse SubscriptionResponse
// 			err = json.Unmarshal(message, &subStatusResponse)
// 			if err != nil {
// 				// TODO: fix
// 				log.Fatal("Unmarshall err1", err)
// 			}
// 			cl.subStatusResponseChannel <- subStatusResponse

// 			select {
// 			case <-ctx.Done():
// 				break Loop
// 			default:
// 			}
// 		}
// 		return
// 	}()
// }

// func (cl *Client) startTickerListener(ctx context.Context, quoteCh chan<- broker.Quote) {
// 	go func() {
// 	Loop:
// 		for {
// 			message, err := cl.readMessage()
// 			if err != nil {
// 				// TODO: fix
// 				log.Println("read error, skipping")
// 				return
// 			}

// 			select {
// 			case <-ctx.Done():
// 				log.Println("Kraken interrupt")
// 				err := cl.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 				if err != nil {
// 					log.Println("write close:", err)
// 					return
// 				}
// 				cl.conn.Close()
// 				break Loop
// 			default:
// 				var tickerResponse TickerResponse
// 				err = json.Unmarshal(message, &tickerResponse)
// 				if err != nil {
// 					// TODO: fix
// 					log.Fatal("Unmarshall err3", err, message)
// 				}
// 				if tickerResponse.Ask != "" {
// 					quoteCh <- *broker.NewExchangeQuote("kraken", cl.channelPairMap[tickerResponse.ChannelID], tickerResponse.Ask)
// 				}
// 			}
// 		}
// 		return
// 	}()
// }
