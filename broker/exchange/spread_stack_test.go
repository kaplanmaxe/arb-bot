package exchange_test

import (
	"testing"

	"github.com/kaplanmaxe/helgart/broker/exchange"
)

var mockSpreads = []exchange.SpreadNode{
	exchange.SpreadNode{
		Price: 8000.01,
		Size:  0.43,
	},
	exchange.SpreadNode{
		Price: 8000.04,
		Size:  0.04,
	},
	exchange.SpreadNode{
		Price: 8000.02,
		Size:  0.243,
	},
	exchange.SpreadNode{
		Price: 8000.90,
		Size:  0.67,
	},
	exchange.SpreadNode{
		Price: 8000.83,
		Size:  0.145,
	},
	exchange.SpreadNode{
		Price: 8000.07,
		Size:  0.98,
	},
	exchange.SpreadNode{
		Price: 8000.13,
		Size:  0.99,
	},
	exchange.SpreadNode{
		Price: 8000.25,
		Size:  0.67,
	},
	exchange.SpreadNode{
		Price: 8000.88,
		Size:  0.76,
	},
	exchange.SpreadNode{
		Price: 8000.33,
		Size:  0.11,
	},
	exchange.SpreadNode{
		Price: 8000.44,
		Size:  0.25,
	},
	exchange.SpreadNode{
		Price: 8000.52,
		Size:  0.43,
	},
	exchange.SpreadNode{
		Price: 8000.56,
		Size:  0.88,
	},
	exchange.SpreadNode{
		Price: 8000.69,
		Size:  0.69,
	},
	exchange.SpreadNode{
		Price: 8000.73,
		Size:  0.43,
	},
}

func TestStackCreation(t *testing.T) {
	stack := exchange.NewSpreadStack(10, "bid")
	if stack.Cap != 10 {
		t.Errorf("Expected size of %d but got size of %d", 10, stack.Cap)
	}

	if stack.Side != "bid" {
		t.Errorf("Expected side of %s but got %s", "bid", stack.Side)
	}
}

func TestStackBidsSorting(t *testing.T) {
	stack := exchange.NewSpreadStack(10, "bid")
	for i := 0; i < 10; i++ {
		stack.Push(&mockSpreads[i])
	}

	highPrice := stack.Nodes[0].Price
	for key, val := range stack.Nodes {
		if val.Price > highPrice {
			t.Errorf("Nodes are not sorted properly. Index %d is greater than index 0", key)
		}
	}
}

func TestStackAsksSorting(t *testing.T) {
	stack := exchange.NewSpreadStack(10, "ask")
	for i := 0; i < 10; i++ {
		stack.Push(&mockSpreads[i])
	}
	highPrice := stack.Nodes[0].Price
	for key, val := range stack.Nodes {
		if val.Price < highPrice {
			t.Errorf("Nodes are not sorted properly. Index %d is less than index 0", key)
		}
	}
}

func TestStackSizeCap(t *testing.T) {
	stack := exchange.NewSpreadStack(10, "bid")
	for i := 0; i < 10; i++ {
		stack.Push(&mockSpreads[i])
	}
	stack.Push(&exchange.SpreadNode{
		Price: 9000,
		Size:  1.5,
	})
	if len(stack.Nodes) != stack.Cap {
		t.Errorf("Stack has a cap of %d but there are %d nodes", stack.Cap, len(stack.Nodes))
	}
	if stack.Nodes[0].Price != 9000 {
		t.Errorf("Stack was not sorted correctly.")
	}
}

func TestStackPop(t *testing.T) {
	stack := exchange.NewSpreadStack(10, "bid")
	for i := 0; i < 10; i++ {
		stack.Push(&mockSpreads[i])
	}
	stack.Pop(8000.01)
	for key, val := range stack.Nodes {
		if val.Price == 8000.01 {
			t.Errorf("The node with a price of %f should have been removed but it is at key %d", 8000.01, key)
		}
	}
	if len(stack.Nodes) != 9 {
		t.Errorf("The stack size should be %d but it is %d", 9, len(stack.Nodes))
	}
}

func TestPricesMap(t *testing.T) {
	stack := exchange.NewSpreadStack(10, "bid")
	for i := 0; i < 10; i++ {
		stack.Push(&mockSpreads[i])
	}
	for _, val := range stack.Nodes {
		if _, ok := stack.Prices[val.Price]; !ok {
			t.Errorf("%f is a node however it does not show up in the map: %#v", val.Price, stack.Prices)
		}
	}
	stack.Pop(8000.01)
	if _, ok := stack.Prices[8000.01]; ok {
		t.Errorf("%f should not be in the prices map but it is there: %#v", 8000.01, stack.Prices)
	}
	stack.Push(&exchange.SpreadNode{
		Price: 9000.00,
		Size:  123,
	})
	if _, ok := stack.Prices[9000.00]; !ok {
		t.Errorf("%f should be in the prices map but it is not there: %#v", 9000.00, stack.Prices)
	}
	// Should just update size
	stack.Push(&exchange.SpreadNode{
		Price: 9000.00,
		Size:  456.1,
	})
	for _, val := range stack.Nodes {
		if val.Price == 9000.00 && val.Size != 456.1 {
			t.Errorf("Node with price %f did not update size correctly. Should have size %f but has size %f", 9000.00, 456.1, val.Size)
		}
	}
}
