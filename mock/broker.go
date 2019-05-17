package mock

import (
	"github.com/kaplanmaxe/helgart/exchange"
)

func MakeMockProductMap() exchange.ProductMap {
	productMap := make(exchange.ProductMap)
	exchanges := []string{exchange.COINBASE, exchange.KRAKEN}
	for _, ex := range exchanges {
		if _, ok := productMap[ex]; !ok {
			productMap[ex] = make(exchange.ExchangeProductMap)
		}
		switch ex {
		case exchange.KRAKEN:
			productMap[exchange.KRAKEN]["MOCK/USD"] = exchange.Product{
				Exchange: "Kraken",
				HePair:   "MOCK-USD",
			}
		case exchange.COINBASE:
			productMap[exchange.COINBASE]["MOCK-USD"] = exchange.Product{
				Exchange: "Kraken",
				HePair:   "MOCK-USD",
			}
		}
	}
	return productMap
}
