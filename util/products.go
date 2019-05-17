package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-yaml/yaml"
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

type product struct {
	ID       int    `json:"id"`
	Exchange string `json:"exchange"`
	ExPair   string `json:"ex_pair"`
	HePair   string `json:"he_pair"`
	ExBase   string `json:"ex_base"`
	ExQuote  string `json:"ex_quote"`
	HeBase   string `json:"he_base"`
	HeQuote  string `json:"he_quote"`
}

type client struct {
	cfg config
	db  *sql.DB
}

type apiResponse struct {
	Data []apiProduct `json:"Data"`
}
type apiProduct struct {
	Exchange string `json:"exchange"`
	ExPair   string `json:"ex_pair"`
	HePair   string `json:"he_pair"`
	ExBase   string `json:"exchange_fsym"`
	ExQuote  string `json:"exchange_tsym"`
	HeBase   string `json:"fsym"`
	HeQuote  string `json:"tsym"`
}

func (c *client) getConfig() error {
	var cfg config
	b, err := ioutil.ReadFile("./.config.yml")
	if err != nil {
		return fmt.Errorf("Error looking for config file: %s", err)
	}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return fmt.Errorf("Error unmarshalling config file: %s", err)
	}
	c.cfg = cfg
	return nil
}

func (c *client) connectDB(dsn string) error {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return fmt.Errorf("Error connecting to db: %s", err)
	}
	if c.db == nil {
		c.db = db

	}

	return nil
}

func (c *client) fetchProducts() ([]product, error) {
	var products []product
	results, err := c.db.Query("SELECT * FROM products")

	if err != nil {
		return products, fmt.Errorf("Error getting products: %s", err)
	}

	for results.Next() {
		var p product

		err = results.Scan(&p.ID, &p.Exchange, &p.ExPair, &p.HePair, &p.ExBase, &p.ExQuote, &p.HeBase, &p.HeQuote)
		if err != nil {
			return products, fmt.Errorf("Error getting products: %s", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (c *client) checkIfExchangeFetched(exchange string) (int, error) {
	row := c.db.QueryRow("SELECT count(*) exchange FROM products WHERE exchange = ?", &exchange)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("Error fetching rows for exchange %s: %s", exchange, err)
	}

	return count, nil
}

// {"exchange":"Kraken","exchange_fsym":"XTZ","exchange_tsym":"ETH","fsym":"XTZ","tsym":"ETH","last_update":1554978692}
// getProducts fetches all products from given api
func (c *client) getProducts(exchange string) ([]apiProduct, error) {
	var response apiResponse
	u := url.URL{Scheme: "https", Host: "min-api.cryptocompare.com", Path: "/data/pair/mapping/exchange"}
	q := u.Query()
	q.Add("e", exchange)
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String())
	if err != nil {
		return []apiProduct{}, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, &response)

	if err != nil {
		return []apiProduct{}, err
	}
	var products []apiProduct
	for _, val := range response.Data {
		var chars []rune
		if val.Exchange == "Kraken" && len(val.ExBase) == 4 && []rune(val.ExBase)[0] == 'X' {
			chars = []rune(val.ExBase)
			val.ExBase = string(chars[1:len(chars)])
		}
		if val.Exchange == "Kraken" && len(val.ExQuote) == 4 && ([]rune(val.ExQuote)[0] == 'X' || []rune(val.ExQuote)[0] == 'Z') {
			chars = []rune(val.ExQuote)
			val.ExQuote = string(chars[1:len(chars)])
		}
		val.HePair = fmt.Sprintf("%s-%s", val.HeBase, val.HeQuote)
		val.ExPair = fmt.Sprintf("%s-%s", val.ExBase, val.ExQuote)
		products = append(products, val)
	}
	return products, nil
}

func (c *client) insertProducts(exchange string, products []apiProduct) error {
	for _, val := range products {
		_, err := c.db.Exec(
			`INSERT INTO products (
				exchange,
				ex_pair,
				he_pair,
				ex_base,
				ex_quote,
				he_base,
				he_quote
			) VALUES (
				?,
				?,
				?,
				?,
				?,
				?,
				?
			);`,
			&val.Exchange,
			&val.ExPair,
			&val.HePair,
			&val.ExBase,
			&val.ExQuote,
			&val.HeBase,
			&val.HeQuote,
		)
		if err != nil {
			return fmt.Errorf("Error inserting products for %s: %s", exchange, err)
		}
	}
	return nil
}

func main() {
	c := &client{}
	err := c.getConfig()

	if err != nil {
		log.Fatal(err)
	}

	err = c.connectDB(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.cfg.DB.Username, c.cfg.DB.Password, c.cfg.DB.Host, c.cfg.DB.Port, c.cfg.DB.DBName))

	defer c.db.Close()

	exchanges := []string{"Binance", "Bitfinex", "Coinbase", "Kraken"}
	for _, val := range exchanges {
		count, err := c.checkIfExchangeFetched(val)
		if err != nil {
			log.Fatal(err)
		}

		// If we don't have the products yet, get them
		if count == 0 {
			products, err := c.getProducts(val)
			if err != nil {
				log.Fatal(err)
			}
			err = c.insertProducts(val, products)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Println("Products inserted!")
}
