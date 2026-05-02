package core

import "testing"

func TestCoreGenesisAndConsensusData(t *testing.T) {
	nextCoreBlockID = 0
	g := GenesisBlock(10, "genesis-meta")
	if g == nil {
		t.Fatal("genesis block is nil")
	}
	if got, want := g.Height(), uint64(0); got != want {
		t.Fatalf("genesis height: got=%d want=%d", got, want)
	}
	if got, want := g.MinterID(), 10; got != want {
		t.Fatalf("genesis minter: got=%d want=%d", got, want)
	}
	if got, want := g.ConsensusData(), any("genesis-meta"); got != want {
		t.Fatalf("consensus data: got=%v want=%v", got, want)
	}
	if _, ok := g.ParentID(); ok {
		t.Fatal("genesis ParentID() should not exist")
	}
}

func TestCoreBlockParentIDAndChainRelation(t *testing.T) {
	nextCoreBlockID = 0
	g := GenesisBlock(1, nil)
	a := NewBlock(g, 2, 10, map[string]any{"difficulty": 123})
	a2 := NewBlock(a, 2, 20, nil)
	b := NewBlock(g, 3, 11, nil)

	if got, ok := a.ParentID(); !ok || got != g.ID() {
		t.Fatalf("a ParentID(): got=(%d,%v) want=(%d,true)", got, ok, g.ID())
	}

	if !a2.IsOnSameChainAs(a) {
		t.Fatal("a2 and a should be on the same chain")
	}
	if a2.IsOnSameChainAs(b) {
		t.Fatal("a2 and b should not be on the same chain")
	}

	if got := a2.BlockWithHeight(0); got != g {
		t.Fatal("BlockWithHeight(0) should return genesis")
	}
}
