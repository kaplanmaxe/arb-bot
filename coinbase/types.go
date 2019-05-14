package coinbase

// {
//     "type": "subscribe",
//     "product_ids": [
//         "ETH-USD",
//         "ETH-EUR"
//     ],
//     "channels": [
//         "heartbeat",
//         {
//             "name": "ticker",
//             "product_ids": [
//                 "ETH-BTC",
//                 "ETH-USD"
//             ]
//         }
//     ]
// }

type subscribeRequest struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []struct {
		Name       string   `json:"name"`
		ProductIDs []string `json:"product_ids"`
	} `json:"channels"`
}

// {
// 	"type":"ticker",
// 	"sequence":6612313774,
// 	"product_id":"ETH-USD",
// 	"price":"176.01000000",
// 	"open_24h":"162.52000000",
// 	"volume_24h":"203511.14306374",
// 	"low_24h":"162.52000000",
// 	"high_24h":"180.69000000",
// 	"volume_30d":"3304909.72266988",
// 	"best_bid":"175.95",
// 	"best_ask":"176.04",
// 	"side":"buy",
// 	"time":"2019-05-07T13:09:07.402000Z",
// 	"trade_id":46871901,
// 	"last_size":"19.17919361"
// }
type tickerResponse struct {
	Pair  string `json:"product_id"`
	Price string `json:"best_ask"`
}

type productsResponse struct {
	Pair          string `json:"id"`
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
}
