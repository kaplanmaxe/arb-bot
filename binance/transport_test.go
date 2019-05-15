package binance_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kaplanmaxe/helgart/binance"
	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/mock"
)

func ignoreFunc(msg []byte) bool {
	return false
}
func TestStart(t *testing.T) {
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()

	client := binance.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	client.Start(ctx)

	mockResponse := []binance.TickerResponse{
		binance.TickerResponse{
			AskQuantity: "123",
			Pair:        "BTCUSD",
			Price:       "1000000.00",
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
			if mockResponse[0].Pair != quote.Pair || mockResponse[0].Price != quote.Price {
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
