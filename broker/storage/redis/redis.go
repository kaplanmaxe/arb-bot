package redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/kaplanmaxe/helgart/broker/exchange"
)

// Client represents a new redis client
type Client struct {
	conn *redis.Client
}

// Config represents a redis config
type Config struct {
	Host string
	Port int
}

// NewClient returns a new redis client
func NewClient(cfg *Config) exchange.ProductCache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &Client{conn: client}
}

// Connect sends a ping to the redis instance to see if we get a result back
func (c *Client) Connect() error {
	_, err := c.conn.Ping().Result()
	if err != nil {
		return fmt.Errorf("Could not connect to redis instance: %s", err)
	}
	return nil
}

// SetPair sets a price of a pair for a given exchange in redis
func (c *Client) SetPair(pair, exchange, price string) error {
	_, err := c.conn.HSet(pair, exchange, price).Result()
	if err != nil {
		return fmt.Errorf("Could not set pair %s in redis: %s", pair, err)
	}
	return nil
}
