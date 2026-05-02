package node

import (
	"testing"

	"github.com/teiyou416/simblock_go/block"
)

func TestReceiveBlockPrefersLongerChain(t *testing.T) {
	n := New(1, 0)
	g := block.Genesis(1)

	if ok := n.ReceiveBlock(g); !ok {
		t.Fatal("expected genesis accepted")
	}

	b1 := block.NewBlock(g, 2, 10)
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
	n := NewWithHashPower(7, 2, 12345)
	if got, want := n.HashPower(), uint64(12345); got != want {
		t.Fatalf("HashPower(): got=%d want=%d", got, want)
	}

	n2 := NewWithHashPower(8, 2, 0)
	if got, want := n2.HashPower(), uint64(1); got != want {
		t.Fatalf("HashPower() minimum fallback: got=%d want=%d", got, want)
	}
}
