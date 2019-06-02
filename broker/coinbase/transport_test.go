package coinbase_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kaplanmaxe/helgart/broker/coinbase"
	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/mock"
)

func ignoreFunc(msg []byte) bool {
	if strings.Contains(string(msg), "product_ids") {
		return true
	}
	return false
}
func TestStart(t *testing.T) {
	quoteCh := make(chan exchange.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()

	client := coinbase.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	productMap := mock.MakeMockProductMap()
	doneCh := make(chan struct{}, 1)
	client.Start(ctx, productMap, doneCh)
	snapshotResponse := &coinbase.SnapshotResponse{
		Type: "snapshot",
		Pair: "MOCK-USD",
		Bids: [][]string{{"1000000", "1"}},
		Asks: [][]string{{"99999.99", "2"}},
	}
	// mockResponse := &coinbase.TickerResponse{
	// 	Pair: "MOCK-USD",
	// 	Bid:  "1000000.00",
	// }
	msg, err := json.Marshal(snapshotResponse)
	if err != nil {
		t.Fatalf("Error marshalling json: %s", err)
	}
	err = client.API.WriteMessage(msg)
	if err != nil {
		t.Fatalf("Error writing message: %s", err)
	}
	mockResponse := &coinbase.LevelTwoResponse{
		Type:    "l2update",
		Pair:    "MOCK-USD",
		Changes: [][]string{{"buy", "1000001.00000000", "3"}},
	}
	msg, err = json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Error marshalling json: %s", err)
	}
	err = client.API.WriteMessage(msg)
	if err != nil {
		t.Fatalf("Error writing message: %s", err)
	}
listener:
	for {
		select {
		case quote := <-quoteCh:
			if mockResponse.Pair != quote.HePair || mockResponse.Changes[0][1] != quote.Bid {
				t.Fatalf("Expecting response %#v but got %#v", mockResponse, quote)
			}
			break listener
		case err := <-errorCh:
			t.Fatalf("%s", err)
			break listener
		default:
		}
	}
	client.API.Close()
}
