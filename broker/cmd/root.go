package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/helgart/broker/api"
	"github.com/kaplanmaxe/helgart/broker/binance"
	"github.com/kaplanmaxe/helgart/broker/bitfinex"
	"github.com/kaplanmaxe/helgart/broker/coinbase"
	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/kraken"
	"github.com/kaplanmaxe/helgart/broker/storage/mysql"
	"github.com/kaplanmaxe/helgart/broker/wsapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow for browsers to connect
}

// Client represents a client connected to the wsapi
type websocketClient struct {
	conn   *websocket.Conn
	sendCh chan *exchange.ArbMarket
}

type websocketAPI struct {
	broker         *exchange.Broker
	quoteCh        chan exchange.Quote
	arbCh          chan *exchange.ArbMarket
	errorCh        chan error
	interruptCh    chan os.Signal
	exchangeDoneCh chan struct{}
	writeCh        chan []byte
	arbMap         map[string]*exchange.ArbMarket
	mtx            *sync.Mutex
	conns          map[*websocket.Conn]*websocketClient
}

func (ws *websocketAPI) quoteHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
loop:
	for {
		select {
		case quote := <-ws.quoteCh:
			err = c.WriteJSON(quote)
			if err != nil {
				break loop
			}
		}
	}
}

func (ws *websocketAPI) arbitrageHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	client := &websocketClient{
		conn:   c,
		sendCh: make(chan *exchange.ArbMarket),
	}
	ws.conns[c] = client
	// c.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// On initial connect we want to send clients all the current opportunites broker has found
	markets, err := ws.marshalArbMarkets()
	if err != nil {
		// TODO: send message back to client
		ws.errorCh <- fmt.Errorf("Error marshalling initial markets: %s", err)
	}
	ws.mtx.Lock()
	err = c.WriteMessage(websocket.BinaryMessage, markets)
	ws.mtx.Unlock()
	// go ws.writePump(c)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	for market := range client.sendCh {
		msg, err := ws.marshalArbMarket(market)
		if err != nil {
			continue
		}
		// TODO: should we care about the error here? I don't think so but maybe
		c.WriteMessage(websocket.BinaryMessage, msg)

		// ws.writeCh <- msg
	}
}

func (ws *websocketAPI) writePump(c *websocket.Conn) {
	for msg := range ws.writeCh {
		ws.mtx.Lock()
		err := c.WriteMessage(websocket.BinaryMessage, msg)
		ws.mtx.Unlock()
		if err != nil {
			break
			// log.Println("Error sending inital markets message")
		}
	}
}

func (ws *websocketAPI) marshalArbMarkets() ([]byte, error) {
	var markets []*wsapi.ArbMarket
	for _, val := range ws.arbMap {
		markets = append(markets, &wsapi.ArbMarket{
			HeBase: val.HeBase,
			Spread: val.Spread,
			Low: &wsapi.ArbMarket_ActiveMarket{
				Exchange:          val.Low.Exchange,
				HePair:            val.Low.HePair,
				ExPair:            val.Low.ExPair,
				Price:             fmt.Sprintf("%8.8f", val.Low.Price),
				TriangulatedPrice: fmt.Sprintf("%8.8f", val.Low.TriangulatedPrice),
			},
			High: &wsapi.ArbMarket_ActiveMarket{
				Exchange:          val.High.Exchange,
				HePair:            val.High.HePair,
				ExPair:            val.High.ExPair,
				Price:             fmt.Sprintf("%8.8f", val.High.Price),
				TriangulatedPrice: fmt.Sprintf("%8.8f", val.High.TriangulatedPrice),
			},
		})
	}
	pb := &wsapi.ArbMarkets{
		Markets: markets,
	}
	return proto.Marshal(pb)
}

func (ws *websocketAPI) marshalArbMarket(market *exchange.ArbMarket) ([]byte, error) {
	pb := &wsapi.ArbMarket{
		HeBase: market.HeBase,
		Spread: market.Spread,
		Low: &wsapi.ArbMarket_ActiveMarket{
			Exchange:          market.Low.Exchange,
			HePair:            market.Low.HePair,
			ExPair:            market.Low.ExPair,
			Price:             fmt.Sprintf("%8.8f", market.Low.Price),
			TriangulatedPrice: fmt.Sprintf("%8.8f", market.Low.TriangulatedPrice),
		},
		High: &wsapi.ArbMarket_ActiveMarket{
			Exchange:          market.High.Exchange,
			HePair:            market.High.HePair,
			ExPair:            market.High.ExPair,
			Price:             fmt.Sprintf("%8.8f", market.High.Price),
			TriangulatedPrice: fmt.Sprintf("%8.8f", market.High.TriangulatedPrice),
		},
	}
	return proto.Marshal(pb)
}

func (ws *websocketAPI) serveWS() {
	http.HandleFunc("/ticker", ws.quoteHandler)
	http.HandleFunc("/arb", ws.arbitrageHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", viper.Get("api.host"), viper.Get("api.port")), nil)
}

func (ws *websocketAPI) connectDB() (exchange.ProductStorage, error) {
	db := mysql.NewClient(&mysql.Config{
		Username: viper.Get("db.helgart_db_username").(string),
		Password: viper.Get("db.helgart_db_password").(string),
		DBName:   viper.Get("db.helgart_db_name").(string),
		Host:     viper.Get("db.helgart_db_host").(string),
		Port:     viper.Get("db.helgart_db_port").(int),
	})
	err := db.Connect()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (ws *websocketAPI) getExchanges() []exchange.Exchange {
	return []exchange.Exchange{
		binance.NewClient(api.NewWebSocketHelper(exchange.BINANCE), ws.quoteCh, ws.errorCh),
		kraken.NewClient(api.NewWebSocketHelper(exchange.KRAKEN), ws.quoteCh, ws.errorCh),
		coinbase.NewClient(api.NewWebSocketHelper(exchange.COINBASE), ws.quoteCh, ws.errorCh),
		bitfinex.NewClient(api.NewWebSocketHelper(exchange.BITFINEX), ws.quoteCh, ws.errorCh),
	}
}

func (ws *websocketAPI) startBroker(ctx context.Context, exchanges []exchange.Exchange, db exchange.ProductStorage) error {
	ws.broker = exchange.NewBroker(exchanges, db)
	return ws.broker.Start(ctx, ws.exchangeDoneCh)
}

func (ws *websocketAPI) startBrokerPump() {
	for {
		select {
		case quote := <-ws.quoteCh:
			if _, ok := ws.broker.ArbProducts[quote.HeBase]; ok {
				// TODO: investigate this bug where coinbase returns no price for MKR-BTC
				if quote.Ask == "" || quote.Bid == "" {
					continue
				}
				bid, err := strconv.ParseFloat(quote.Bid, 64)
				if err != nil {
					ws.errorCh <- err
				}
				ask, err := strconv.ParseFloat(quote.Ask, 64)
				if err != nil {
					ws.errorCh <- err
				}
				// Store the old high and low so we only send messages back to clients if a market has changed
				pair := ws.broker.ActiveMarkets[quote.HeBase]
				var oldHigh, oldLow float64
				if pair != nil && len(pair.Bids) > 0 && len(pair.Asks) > 0 {
					oldHigh = pair.Bids[0].TriangulatedPrice
					oldLow = pair.Asks[0].TriangulatedPrice
				}
				ws.broker.InsertActiveMarket(&exchange.ActiveMarket{
					Exchange: quote.Exchange,
					HePair:   quote.HePair,
					ExPair:   quote.ExPair,
					HeBase:   quote.HeBase,
					HeQuote:  quote.HeQuote,
					Bid:      bid,
					Ask:      ask,
				})
				// Not sure why this is here?
				if oldHigh == 0 && oldLow == 0 {
					continue
				}
				// If there is more than one quote and the best bid or ask has changed, we perform an update operation
				if len(ws.broker.ActiveMarkets[quote.HeBase].Bids) > 1 && len(ws.broker.ActiveMarkets[quote.HeBase].Asks) > 1 &&
					oldHigh != pair.Bids[0].TriangulatedPrice && oldLow != pair.Bids[0].TriangulatedPrice {
					high := pair.Bids[0] // Sell at highest price
					low := pair.Asks[0]  // Buy at lowest price
					if market := exchange.NewArbMarket(quote.HeBase, low, high); market.Spread >= 0.01 {
						ws.arbMap[market.HeBase] = market
						for _, client := range ws.conns {
							if _, ok := ws.conns[client.conn]; ok {
								client.sendCh <- market
							} else {
								delete(ws.conns, client.conn)
								close(client.sendCh)
							}

						}
						// ws.arbCh <- market
					}

				}
			}
		case err := <-ws.errorCh:
			// TODO: send a message back to the client
			log.Println(err)
			break
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "start",
	Short: "Start starts the broker service",
	Long:  `broker fetches cryptocurrency markets and potentially exposes a websocket API`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Can't read config: %s", err)
			os.Exit(1)
		}
		ws := &websocketAPI{
			quoteCh:        make(chan exchange.Quote),
			arbCh:          make(chan *exchange.ArbMarket),
			errorCh:        make(chan error),
			interruptCh:    make(chan os.Signal, 1),
			exchangeDoneCh: make(chan struct{}),
			writeCh:        make(chan []byte),
			arbMap:         make(map[string]*exchange.ArbMarket),
			mtx:            &sync.Mutex{},
			conns:          make(map[*websocket.Conn]*websocketClient),
		}
		// Start websocket API
		go ws.serveWS()
		// Connect to db
		db, err := ws.connectDB()
		if err != nil {
			log.Fatal(err)
		}

		// Interrupt handler logic
		log.Print("Starting quote server")
		doneCh := make(chan struct{}, 1)
		signal.Notify(ws.interruptCh, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())

		// Start broker
		exchanges := ws.getExchanges()
		err = ws.startBroker(ctx, exchanges, db)
		if err != nil {
			log.Fatal(err)
		}

		ws.startBrokerPump()
		// interrupt handler
		go func() {
			<-ws.interruptCh
			log.Println("interrupt received")
			cancel()
			canceledExchanges := 0
		interrupt:
			for {
				select {
				case <-ws.exchangeDoneCh:
					canceledExchanges++
					if canceledExchanges == len(exchanges) {
						close(doneCh)
						break interrupt
					}
				}
			}

		}()
		<-doneCh
	},
}

var cfgFile string

// Execute is the main entry point to the application and executes the root cmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.helgart/.broker.config.yml)")
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("yaml")
		viper.SetConfigName("broker.config")
		viper.AddConfigPath("$HOME/.helgart")
	}
}
