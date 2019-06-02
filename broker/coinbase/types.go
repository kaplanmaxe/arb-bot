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

// TickerResponse is a response from coinbase api
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
type TickerResponse struct {
	Pair string `json:"product_id"`
	Ask  string `json:"best_ask"`
	Bid  string `json:"best_bid"`
}

// SnapshotResponse is the intial response coinbase sends to you on the level2 channel
type SnapshotResponse struct {
	Type string     `json:"type"`
	Pair string     `json:"product_id"`
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

// LevelTwoResponse is a response from the level2 channel on coinbase's api
// TODO: do all these types have to be exported?
type LevelTwoResponse struct {
	Type    string     `json:"type"`
	Pair    string     `json:"product_id"`
	Changes [][]string `json:"changes"`
}

type productsResponse struct {
	Pair          string `json:"id"`
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
}
