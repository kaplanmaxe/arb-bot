package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/kaplanmaxe/helgart/broker/api"
	"github.com/kaplanmaxe/helgart/broker/binance"
	"github.com/kaplanmaxe/helgart/broker/bitfinex"
	"github.com/kaplanmaxe/helgart/broker/coinbase"
	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/kraken"
	"github.com/kaplanmaxe/helgart/broker/storage/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type websocketHandler struct {
	quoteCh     chan exchange.Quote
	errorCh     chan error
	interruptCh chan os.Signal
}

func (ws *websocketHandler) quoteHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// select {
	// case quote := <-ws.quoteCh:
	// 	// fmt.Println(1)
	// 	err = c.WriteJSON(quote)
	// 	if err != nil {
	// 		fmt.Println("Closed")
	// 		c.Close()
	// 		// break handler
	// 	}
	// default:
	// 	fmt.Println("OH")
	// }
	// quote := <-ws.quoteCh
	// err = c.WriteJSON(quote)
	// if err != nil {
	// 	fmt.Println("WS Conn closed")
	// 	c.Close()
	// }
	defer c.Close()
	// handler:
	// 	for {
	// 		select {
	// 		case quote := <-ws.quoteCh:
	// 			// fmt.Println(1)
	// 			err = c.WriteJSON(quote)
	// 			if err != nil {
	// 				fmt.Println("Closed")
	// 				// c.Close()
	// 				break handler
	// 			}
	// 		// case <-ws.interruptCh:
	// 		// 	// fmt.Println(3)
	// 		// 	c.Close()
	// 		// 	fmt.Println("NOW")
	// 		// 	break handler
	// 		default:
	// 			// fmt.Println(2)
	// 		}

	// 	}
	for quote := range ws.quoteCh {
		err = c.WriteJSON(quote)
		if err != nil {
			fmt.Println("Closed")
			// c.Close()
			break
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "start",
	Short: "Start starts the broker service",
	Long:  `broker fetches cryptocurrency markets and potentially exposes a websocket API`,
	Run: func(cmd *cobra.Command, args []string) {
		ws := &websocketHandler{
			quoteCh:     make(chan exchange.Quote),
			interruptCh: make(chan os.Signal, 1),
		}
		go func() {
			http.HandleFunc("/", ws.quoteHandler)
			http.ListenAndServe("localhost:8080", nil)
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
		// interrupt := make(chan os.Signal, 1)
		doneCh := make(chan struct{}, 1)
		// quoteCh := make(chan exchange.Quote)
		errorCh := make(chan error)

		signal.Notify(ws.interruptCh, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())

		broker := exchange.NewBroker([]exchange.Exchange{
			kraken.NewClient(api.NewWebSocketHelper(exchange.KRAKEN), ws.quoteCh, errorCh),
			coinbase.NewClient(api.NewWebSocketHelper(exchange.COINBASE), ws.quoteCh, errorCh),
			binance.NewClient(api.NewWebSocketHelper(exchange.BINANCE), ws.quoteCh, errorCh),
			bitfinex.NewClient(api.NewWebSocketHelper(exchange.BITFINEX), ws.quoteCh, errorCh),
		}, db)
		err = broker.Start(ctx)

		if err != nil {
			log.Fatal(err)
		}

		go func() {
			<-ws.interruptCh
			log.Println("interrupt received")
			cancel()
			close(doneCh)
		}()
		<-doneCh
	},
}

func arbitrageHandler(broker *exchange.Broker, quote exchange.Quote) {
	if _, ok := broker.ArbProducts[quote.HePair]; ok {
		// TODO: investigate this bug where coinbase returns no price for MKR-BTC
		if quote.Price == "" {
			return
		}
		price, err := strconv.ParseFloat(quote.Price, 64)
		if err != nil {
			log.Fatal(err)
		}
		quote.PriceFloat = price
		broker.InsertActiveMarket(&exchange.ActiveMarket{
			Exchange: quote.Exchange,
			HePair:   quote.HePair,
			ExPair:   quote.ExPair,
			Price:    price,
		})
		var arbMarket exchange.ArbMarket
		if len(broker.ActiveMarkets[quote.HePair]) > 1 {
			pair := broker.ActiveMarkets[quote.HePair]
			low := pair[len(pair)-1]
			high := pair[0]
			arbMarket = *exchange.NewArbMarket(low.HePair, low, high)
			log.Printf("Arb Market: %#v", arbMarket)
		}
	}
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

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Can't read config: %s", err)
		os.Exit(1)
	}
}

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
