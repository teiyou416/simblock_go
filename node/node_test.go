package node_test

import (
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node"
)

func TestReceiveBlockPrefersLongerChain(t *testing.T) {
	n := node.New(1, 0)
	g := core.GenesisBlock(1, nil)

	if ok := n.ReceiveBlock(g); !ok {
		t.Fatal("expected genesis accepted")
	}

	b1 := core.NewBlock(g, 2, 10, nil)
	if ok := n.ReceiveBlock(b1); !ok {
		t.Fatal("expected higher block accepted")
	}
	if n.Tip() != b1 {
		t.Fatal("tip should be b1")
	}

	shorter := g
	if ok := n.ReceiveBlock(shorter); ok {
		t.Fatal("expected shorter block rejected")
	}
	if n.Tip() != b1 {
		t.Fatal("tip should still be b1")
	}
}

func TestNodeHashPower(t *testing.T) {
	n := node.NewWithHashPower(7, 2, 12345)
	if got, want := n.HashPower(), uint64(12345); got != want {
		t.Fatalf("HashPower(): got=%d want=%d", got, want)
	}

	n2 := node.NewWithHashPower(8, 2, 0)
	if got, want := n2.HashPower(), uint64(1); got != want {
		t.Fatalf("HashPower() minimum fallback: got=%d want=%d", got, want)
	}
}

func TestInvRecBlockBroadcastFlow(t *testing.T) {
	timer := engine.NewTimer()
	net := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)

	a := node.NewWithHashPower(1, 0, 1000)
	b := node.NewWithHashPower(2, 1, 1000)
	a.BindEnvironment(timer, net)
	b.BindEnvironment(timer, net)
	a.AddNeighbor(b)
	b.AddNeighbor(a)

	genesis := core.GenesisBlock(a.ID(), nil)
	a.ReceiveBlock(genesis)
	b.ReceiveBlock(genesis)

	child := core.NewBlock(a.Tip(), a.ID(), timer.CurrentTime(), nil)
	a.ReceiveBlock(child)

	timer.RunUntilEmpty()

	if b.Tip() == nil {
		t.Fatal("receiver tip is nil")
	}
	if got, want := b.Tip().ID(), child.ID(); got != want {
		t.Fatalf("receiver tip id: got=%d want=%d", got, want)
	}
}
