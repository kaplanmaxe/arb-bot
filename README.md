# helgart

[![Build Status](https://travis-ci.org/kaplanmaxe/helgart.svg?branch=master)](https://travis-ci.org/kaplanmaxe/helgart)

Helgart is a microservice application for algorithmic trading of cryptocurrencies. **It is still a WIP.**

### Services

| Name | Description |
| -----| ------------|
| broker | fetch prices from exchanges and calculate opportunities for arbitrage |
| mysql | for persistent storage of normalized pair names (Ex: IOT-USD vs IOTA-USD) |

### Connected Crypto Exchanges

- Binance
- Bitfinex
- Coinbase
- Kraken

More to come soon ...

### Installation

`make && make install`

### Running

- Copy `.config.dist.yml` to `.config.yml` and correct values
- `broker --config <PATH_TO_CONFIG>`

### Running with Docker

You can run helgart with docker compose simply with:

`docker-compose up -d`

You can also run it with:

```
make up
make down
```

`make down` will stop helgart