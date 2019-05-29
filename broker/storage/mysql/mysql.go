package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/kaplanmaxe/helgart/broker/exchange"
)

// Config represents a mysql config
type Config struct {
	Username string `yaml:"HELGART_DB_USERNAME"`
	Password string `yaml:"HELGART_DB_PASSWORD"`
	DBName   string `yaml:"HELGART_DB_NAME"`
	Port     int    `yaml:"HELGART_DB_PORT"`
	Host     string `yaml:"HELGART_DB_HOST"`
}

// Client represents a mysql client
type Client struct {
	DB  *sql.DB
	cfg *Config
}

// NewClient returns a new mysql client
func NewClient(cfg *Config) exchange.ProductStorage {
	return &Client{
		cfg: cfg,
	}
}

// Connect opens a connection to the db
func (c *Client) Connect() error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.cfg.Username, c.cfg.Password, c.cfg.Host, c.cfg.Port, c.cfg.DBName))

	if err != nil {
		return fmt.Errorf("Error connecting to db: %s", err)
	}
	if c.DB == nil {
		c.DB = db

	}

	return nil
}

// FetchAllProducts fetch all products from the db so we can normalize pair names
func (c *Client) FetchAllProducts() ([]exchange.Product, error) {
	var products []exchange.Product
	results, err := c.DB.Query("SELECT exchange, ex_pair, he_pair, ex_base, ex_quote, he_base, he_quote FROM products")

	if err != nil {
		return products, fmt.Errorf("Error getting products: %s", err)
	}

	for results.Next() {
		var p exchange.Product

		err = results.Scan(&p.Exchange, &p.ExPair, &p.HePair, &p.ExBase, &p.ExQuote, &p.HeBase, &p.HeQuote)
		if err != nil {
			return products, fmt.Errorf("Error getting products: %s", err)
		}
		products = append(products, p)
	}
	return products, nil
}

// FetchArbProducts fetches all products that have more than one market to arb
func (c *Client) FetchArbProducts() (exchange.ArbProductMap, error) {
	products := make(exchange.ArbProductMap)
	results, err := c.DB.Query("SELECT he_base FROM products GROUP BY he_base HAVING COUNT(he_base) > 1;")

	if err != nil {
		return products, fmt.Errorf("Error getting products: %s", err)
	}

	for results.Next() {
		var p string

		err = results.Scan(&p)
		if err != nil {
			return products, fmt.Errorf("Error getting products: %s", err)
		}
		products[p] = struct{}{}
	}
	return products, nil
}
