package mock

import (
	"github.com/kaplanmaxe/helgart/exchange"
)

func MakeMockProductMap() exchange.ProductMap {
	productMap := make(exchange.ProductMap)
	exchanges := []string{exchange.COINBASE, exchange.KRAKEN, exchange.BINANCE, exchange.BITFINEX}
	for _, ex := range exchanges {
		if _, ok := productMap[ex]; !ok {
			productMap[ex] = make(exchange.ExchangeProductMap)
		}
		switch ex {
		case exchange.KRAKEN:
			productMap[exchange.KRAKEN]["MOCK/USD"] = exchange.Product{
				Exchange: exchange.KRAKEN,
				HePair:   "MOCK-USD",
			}
		case exchange.COINBASE:
			productMap[exchange.COINBASE]["MOCK-USD"] = exchange.Product{
				Exchange: exchange.COINBASE,
				HePair:   "MOCK-USD",
			}
		case exchange.BINANCE:
			productMap[exchange.BINANCE]["MOCKUSD"] = exchange.Product{
				Exchange: exchange.BINANCE,
				HePair:   "MOCK-USD",
				ExPair:   "MOCKUSD",
			}
		case exchange.BITFINEX:
			productMap[exchange.BITFINEX]["MOCKUSD"] = exchange.Product{
				Exchange: exchange.BITFINEX,
				HePair:   "MOCK-USD",
				ExPair:   "MOCKUSD",
			}
		}
	}
	return productMap
}
