package coinbase_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/coinbase"
	"github.com/kaplanmaxe/helgart/mock"
)

func ignoreFunc(msg []byte) bool {
	return false
}
func TestStart(t *testing.T) {
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()

	client := coinbase.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	client.Start(ctx)

	mockResponse := &coinbase.TickerResponse{
		Pair:  "BTCUSD",
		Price: "1000000.00",
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
			if mockResponse.Pair != quote.Pair || mockResponse.Price != quote.Price {
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
