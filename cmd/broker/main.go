package main

import (
	"context"
	"flag"
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
	"github.com/kaplanmaxe/helgart/broker/storage/redis"
)

type db struct {
	Username string `yaml:"HELGART_DB_USERNAME"`
	Password string `yaml:"HELGART_DB_PASSWORD"`
	DBName   string `yaml:"HELGART_DB_NAME"`
	Port     int    `yaml:"HELGART_DB_PORT"`
	Host     string `yaml:"HELGART_DB_HOST"`
}

type cache struct {
	Host string `yaml:"HELGART_CACHE_HOST"`
	Port int    `yaml:"HELGART_CACHE_PORT"`
}

type config struct {
	Version string `yaml:"version"`
	DB      db
	Cache   cache
}

func getConfig(path *string) (config, error) {
	var cfg config
	b, err := ioutil.ReadFile(*path)
	if err != nil {
		return cfg, fmt.Errorf("Error looking for config file: %s", err)
	}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("Error unmarshalling config file: %s", err)
	}

	return cfg, nil
}

const version = "0.0.1-alpha"

func main() {
	var versionCalled bool
	cfgPath := flag.String("config", ".config.yml", "absolute path to config file")
	flag.BoolVar(&versionCalled, "version", false, "version number")
	flag.Parse()
	// version flag
	if versionCalled {
		fmt.Printf("helgart-broker version: %s\n", version)
		os.Exit(0)
	}
	cfg, err := getConfig(cfgPath)
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

	cache := redis.NewClient(&redis.Config{
		Host: cfg.Cache.Host,
		Port: cfg.Cache.Port,
	})
	err = cache.Connect()
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
	}, db, cache)
	err = broker.Start(ctx)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case quote := <-quoteCh:
				if _, ok := broker.ArbProducts[quote.HePair]; ok {
					// TODO: investigate this bug where coinbase returns no price for MKR-BTC
					if quote.Price == "" {
						continue
					}
					cache.SetPair(quote.HePair, quote.Exchange, quote.Price)
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
}
