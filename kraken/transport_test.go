package kraken_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kaplanmaxe/helgart/broker"
	"github.com/kaplanmaxe/helgart/kraken"
	"github.com/kaplanmaxe/helgart/mock"
)

func ignoreFunc(msg []byte) bool {
	if strings.Contains(string(msg), "{\"event\":") {
		return true
	} else {
		return false
	}
}
func TestStart(t *testing.T) {
	quoteCh := make(chan broker.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()
	client := kraken.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	client.Start(ctx)
	for i := 0; i < len(client.Pairs)+1; i++ {
		var channelID int
		if i == 0 {
			channelID = 392
		} else {
			channelID = i
		}
		subscribeResponse := &kraken.SubscriptionResponse{
			ChannelID: channelID,
			Pair:      "MOCKUSD",
		}
		msg, err := json.Marshal(subscribeResponse)
		if err != nil {
			t.Fatalf("Error marshaling subscribe response: %s", err)
		}
		err = client.API.WriteMessage(msg)
		if err != nil {
			t.Fatal(err)
		}
	}
	mockResponse := json.RawMessage("[392,{\"a\":[\"0.43088000\",1200,\"1200.00000000\"],\"b\":[\"0.42950000\",28067,\"28067.80000000\"],\"c\":[\"0.42950000\",\"2000.00000000\"],\"v\":[\"15365184.37786758\",\"116236572.40419129\"],\"p\":[\"0.42806190\",\"0.39588510\"],\"t\":[4403,38241],\"l\":[\"0.40690000\",\"0.34375000\"],\"h\":[\"0.44800000\",\"0.44800000\"],\"o\":[\"0.40820000\",\"0.35385000\"]}]")
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
			if quote.Price != "0.43088000" {
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
