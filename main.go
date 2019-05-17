package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/kaplanmaxe/helgart/broker/api"
	"github.com/kaplanmaxe/helgart/broker/binance"
	"github.com/kaplanmaxe/helgart/broker/bitfinex"
	"github.com/kaplanmaxe/helgart/broker/coinbase"
	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/kraken"
	"github.com/kaplanmaxe/helgart/broker/storage/mysql"
)

type db struct {
	Username string `yaml:"HELGART_DB_USERNAME"`
	Password string `yaml:"HELGART_DB_PASSWORD"`
	DBName   string `yaml:"HELGART_DB_NAME"`
	Port     int    `yaml:"HELGART_DB_PORT"`
	Host     string `yaml:"HELGART_DB_HOST"`
}

type config struct {
	Version string `yaml:"version"`
	DB      db
}

func getConfig() (config, error) {
	var cfg config
	b, err := ioutil.ReadFile("./.config.yml")
	if err != nil {
		return cfg, fmt.Errorf("Error looking for config file: %s", err)
	}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("Error unmarshalling config file: %s", err)
	}

	return cfg, nil
}

func main() {
	cfg, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	db := mysql.NewClient(&mysql.Config{
		Username: cfg.DB.Username,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.DBName,
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
	})
	err = db.Connect()
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
				log.Printf("Quote: %#v", quote)
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
}
