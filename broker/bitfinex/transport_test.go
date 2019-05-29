package bitfinex_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kaplanmaxe/helgart/broker/bitfinex"
	"github.com/kaplanmaxe/helgart/broker/exchange"
	"github.com/kaplanmaxe/helgart/broker/mock"
)

type mockSubscriptionResponse struct {
	ChannelID int    `json:"chanId"`
	Pair      string `json:"pair"`
	OmitMock  bool   `json:"omitmock"`
}

func ignoreFunc(msg []byte) bool {
	if strings.Contains(string(msg), `"symbol"`) {
		return true
	} else {
		return false
	}
}

func TestStart(t *testing.T) {
	quoteCh := make(chan exchange.Quote)
	errorCh := make(chan error, 1)
	ctx := context.TODO()
	client := bitfinex.NewClient(mock.NewConnector(ignoreFunc), quoteCh, errorCh)
	productMap := mock.MakeMockProductMap()
	doneCh := make(chan struct{}, 1)
	client.Start(ctx, productMap, doneCh)
	for i := 0; i < len(client.Pairs)+1; i++ {
		var channelID int
		if i == 0 {
			channelID = 368172

		} else {
			channelID = i
		}
		subscribeResponse := &mockSubscriptionResponse{
			ChannelID: channelID,
			Pair:      "MOCKUSD",
			OmitMock:  true,
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
	mockResponse := json.RawMessage("[368172,[8114.8,67.33192073,8114.9,3.3117539600000003,314.9,0.0404,8114.9,18884.37269102,8199.9,7731]]")
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
			if quote.Ask != "8114.900000" {
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
