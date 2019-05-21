package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

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

var rootCmd = &cobra.Command{
	Use:   "start",
	Short: "Start starts the broker service",
	Long:  `broker fetches cryptocurrency markets and potentially exposes a websocket API`,
	Run: func(cmd *cobra.Command, args []string) {
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
		interrupt := make(chan os.Signal, 1)
		doneCh := make(chan struct{}, 1)
		quoteCh := make(chan exchange.Quote)
		errorCh := make(chan error)

		signal.Notify(interrupt, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())

		broker := exchange.NewBroker([]exchange.Exchange{
			kraken.NewClient(api.NewWebSocketHelper(exchange.KRAKEN), quoteCh, errorCh),
			coinbase.NewClient(api.NewWebSocketHelper(exchange.COINBASE), quoteCh, errorCh),
			binance.NewClient(api.NewWebSocketHelper(exchange.BINANCE), quoteCh, errorCh),
			bitfinex.NewClient(api.NewWebSocketHelper(exchange.BITFINEX), quoteCh, errorCh),
		}, db)
		err = broker.Start(ctx)

		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for {
				select {
				case quote := <-quoteCh:
					// if cfg.Trading.Arbitrage {
					if viper.Get("trading.helgart_arbitrage").(bool) {
						arbitrageHandler(broker, quote)
					} else {
						log.Printf("Quote: %#v", quote)
					}

				case err := <-errorCh:
					fmt.Printf("Error: %s\n", err)
				case <-interrupt:
					log.Println("interrupt received")
					cancel()
					select {
					// TODO: race condition. fix
					case <-time.After(7 * time.Second):
						close(doneCh)
						return
					}
				default:
				}
			}
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
	rootCmd.AddCommand(versionCmd)
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
