package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

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

type websocketAPI struct {
	broker         *exchange.Broker
	quoteCh        chan exchange.Quote
	arbCh          chan *exchange.ArbMarket
	errorCh        chan error
	interruptCh    chan os.Signal
	exchangeDoneCh chan struct{}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow for browsers to connect
}

func (ws *websocketAPI) quoteHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		select {
		case quote := <-ws.quoteCh:
			err = c.WriteJSON(quote)
			if err != nil {
				fmt.Println("Closed")
				break
			}
		case err := <-ws.errorCh:
			log.Println(err)
			break
		}
	}
}

func (ws *websocketAPI) arbitrageHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for market := range ws.arbCh {
		msg, err := ws.pbMarshalArbMarket(market)
		if err != nil {
			continue
		}

		err = c.WriteMessage(websocket.BinaryMessage, msg)
		// msg, err := json.Marshal(market)
		// err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("Closed")
			break
		}
	}
}

func (ws *websocketAPI) pbMarshalArbMarket(market *exchange.ArbMarket) ([]byte, error) {
	pb := &wsapi.ArbMarket{
		HeBase: market.HeBase,
		Spread: market.Spread,
		Low: &wsapi.ArbMarket_ActiveMarket{
			Exchange:          market.Low.Exchange,
			HePair:            market.Low.HePair,
			ExPair:            market.Low.ExPair,
			Price:             fmt.Sprintf("%f", market.Low.Price),
			TriangulatedPrice: fmt.Sprintf("%f", market.Low.TriangulatedPrice),
		},
		High: &wsapi.ArbMarket_ActiveMarket{
			Exchange:          market.High.Exchange,
			HePair:            market.High.HePair,
			ExPair:            market.High.ExPair,
			Price:             fmt.Sprintf("%f", market.High.Price),
			TriangulatedPrice: fmt.Sprintf("%f", market.High.TriangulatedPrice),
		},
	}
	return proto.Marshal(pb)
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
		}
		go func() {
			http.HandleFunc("/ticker", ws.quoteHandler)
			http.HandleFunc("/arb", ws.arbitrageHandler)
			http.ListenAndServe(fmt.Sprintf("%s:%d", viper.Get("api.host"), viper.Get("api.port")), nil)
		}()

		db := mysql.NewClient(&mysql.Config{
			Username: viper.Get("db.helgart_db_username").(string),
			Password: viper.Get("db.helgart_db_password").(string),
			DBName:   viper.Get("db.helgart_db_name").(string),
			Host:     viper.Get("db.helgart_db_host").(string),
			Port:     viper.Get("db.helgart_db_port").(int),
		})
		err := db.Connect()
		if err != nil {
			log.Fatal(err)
		}

		log.Print("Starting quote server")
		doneCh := make(chan struct{}, 1)
		signal.Notify(ws.interruptCh, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())
		exchanges := []exchange.Exchange{
			binance.NewClient(api.NewWebSocketHelper(exchange.BINANCE), ws.quoteCh, ws.errorCh),
			kraken.NewClient(api.NewWebSocketHelper(exchange.KRAKEN), ws.quoteCh, ws.errorCh),
			coinbase.NewClient(api.NewWebSocketHelper(exchange.COINBASE), ws.quoteCh, ws.errorCh),
			bitfinex.NewClient(api.NewWebSocketHelper(exchange.BITFINEX), ws.quoteCh, ws.errorCh),
		}
		ws.broker = exchange.NewBroker(exchanges, db)
		err = ws.broker.Start(ctx, ws.exchangeDoneCh)

		if err != nil {
			log.Fatal(err)
		}

		// quoteHandler
		for quote := range ws.quoteCh {
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
				// quote.PriceFloat = price
				ws.broker.InsertActiveMarket(&exchange.ActiveMarket{
					Exchange: quote.Exchange,
					HePair:   quote.HePair,
					ExPair:   quote.ExPair,
					HeBase:   quote.HeBase,
					HeQuote:  quote.HeQuote,
					Bid:      bid,
					Ask:      ask,
				})
				// var arbMarket exchange.ArbMarket
				if len(ws.broker.ActiveMarkets[quote.HeBase].Bids) > 1 && len(ws.broker.ActiveMarkets[quote.HeBase].Asks) > 1 {
					pair := ws.broker.ActiveMarkets[quote.HeBase]
					high := pair.Bids[0] // Sell at highest price
					low := pair.Asks[0]  // Buy at lowest price
					ws.arbCh <- exchange.NewArbMarket(quote.HeBase, low, high)
				}
			}
		}

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
