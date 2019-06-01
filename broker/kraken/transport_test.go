package kraken_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/kraken"
	"github.com/kaplanmaxe/helgart/broker/mock"
)

func ignoreFunc(msg []byte) bool {
	if strings.Contains(string(msg), "{\"event\":") || strings.Contains(string(msg), "heartbeat") ||
		strings.Contains(string(msg), "connectionID") || strings.Contains(string(msg), "channelID") {
		return true
	} else {
		return false
	}
}
func TestStart(t *testing.T) {
	quoteCh := make(chan exchange.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()

	client := kraken.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	productMap := mock.MakeMockProductMap()
	doneCh := make(chan struct{}, 1)
	client.Start(ctx, productMap, doneCh)
	mockResponse := json.RawMessage(`[835,["0.00555670","0.00564450","1559355056.098606","245.68795774","292.87831705"],"spread","MOCK/USD"]`)
	msg, err := json.Marshal(&mockResponse)
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
			if quote.Ask != "0.00564450" {
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
