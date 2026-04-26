package block

import "testing"

func TestGenesisAndChildHeight(t *testing.T) {
	genesis := Genesis(1)
	if genesis == nil {
		t.Fatal("genesis is nil")
	}
	if genesis.Height() != 0 {
		t.Fatalf("genesis height: got=%d want=0", genesis.Height())
	}

	child := NewBlock(genesis, 2, 10)
	if child.Height() != 1 {
		t.Fatalf("child height: got=%d want=1", child.Height())
	}
	if child.Parent() != genesis {
		t.Fatal("child parent mismatch")
	}
}

func TestIsOnSameChainAs(t *testing.T) {
	g := Genesis(1)
	a := NewBlock(g, 1, 10)
	b := NewBlock(g, 2, 11)
	a2 := NewBlock(a, 1, 20)

	if !a2.IsOnSameChainAs(a) {
		t.Fatal("expected a2 and a on same chain")
	}
	if a2.IsOnSameChainAs(b) {
		t.Fatal("expected a2 and b on different chains")
	}
}
