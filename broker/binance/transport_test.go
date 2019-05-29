package binance_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kaplanmaxe/helgart/broker/binance"
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

	client := binance.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	productMap := mock.MakeMockProductMap()
	doneCh := make(chan struct{}, 1)
	client.Start(ctx, productMap, doneCh)

	mockResponse := []binance.TickerResponse{
		binance.TickerResponse{
			AskQuantity: "123",
			Pair:        "MOCKUSD",
			Bid:         "1000000.00",
		},
	}
	msg, err := json.Marshal(mockResponse)
	if err != nil {
		t.Fatalf("Error marshalling json: %s", err)
	}
	err = client.API.WriteMessage(msg) // test
	if err != nil {
		t.Fatalf("Error writing message: %s", err)
	}
listener:
	for {
		select {
		case quote := <-quoteCh:
			if mockResponse[0].Pair != quote.ExPair || mockResponse[0].Bid != quote.Bid {
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
