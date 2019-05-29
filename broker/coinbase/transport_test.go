package coinbase_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kaplanmaxe/helgart/broker/coinbase"
	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/mock"
)

func ignoreFunc(msg []byte) bool {
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

	mockResponse := &coinbase.TickerResponse{
		Pair: "MOCK-USD",
		Bid:  "1000000.00",
	}
	msg, err := json.Marshal(mockResponse)
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
			if mockResponse.Pair != quote.HePair || mockResponse.Bid != quote.Bid {
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
