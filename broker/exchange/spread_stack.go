package exchange

import (
	"sort"
)

// SpreadStack is a stack to maintain the order book. SpreadStack is a bit different than
// a tradtional stack in the sense that it is sorted. The last node that is pushed into
// the stack has no guarantee that it will be the first element in the stack. Ordering
// will always be sorted after a call to Push(). SpreadStacks maintains a fixed length.
// It also does not follow a linked list approach and all elements in the stack can be
// accessed at a given index. If the market side is ask, we sort ascending and descending
// if bids
type SpreadStack struct {
	Cap    int
	Nodes  []SpreadNode
	Side   string
	Prices map[float64]struct{}
}

// SpreadNode represents a node in a SpreadStack. This will contain a price and size
type SpreadNode struct {
	Price float64
	Size  float64
}

// NewSpreadStack returns a new spread stack
func NewSpreadStack(cap int, side string) *SpreadStack {
	return &SpreadStack{
		Side:   side,
		Nodes:  []SpreadNode{},
		Cap:    cap,
		Prices: make(map[float64]struct{}),
	}
}

// Push pushes an element into the stack and then resorts
func (s *SpreadStack) Push(node *SpreadNode) {
	// If the price is not in prices map stack is full, we first remove the last node (it is sorted at this point),
	// replace it with the new node, and then resort. If it's not full, we just add it.
	if _, ok := s.Prices[node.Price]; !ok {
		if len(s.Nodes) == s.Cap {
			s.Nodes[s.Cap-1] = *node
		} else {
			s.Nodes = append(s.Nodes, *node)
		}
		s.Prices[node.Price] = struct{}{}
	} else {
		// If we have the price, we simply just update size
		for key, val := range s.Nodes {
			if val.Price == node.Price {
				s.Nodes[key].Size = node.Size
			}
		}
	}

	s.sort()
}

// Pop removes a node out of the stack and resorts
func (s *SpreadStack) Pop(price float64) {
	// If we don't have that price, do nothing
	if _, ok := s.Prices[price]; !ok {
		return
	}
	delete(s.Prices, price)
	for key, val := range s.Nodes {
		if val.Price == price {
			s.Nodes[key] = s.Nodes[len(s.Nodes)-1]
			s.Nodes = s.Nodes[:len(s.Nodes)-1]
			break
		}
	}
	s.sort()
}

func (s *SpreadStack) sort() {
	if s.Side == "ask" {
		sort.Slice(s.Nodes, func(i, j int) bool {
			return s.Nodes[i].Price < s.Nodes[j].Price
		})
	} else if s.Side == "bid" {
		sort.Slice(s.Nodes, func(i, j int) bool {
			return s.Nodes[i].Price > s.Nodes[j].Price
		})
	}

}
