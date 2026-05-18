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

func TestCoreBlockUnclesAreStoredAndCopied(t *testing.T) {
	nextCoreBlockID = 0
	g := GenesisBlock(1, nil)
	a := NewBlock(g, 2, 10, nil)
	u1 := NewBlock(g, 3, 11, nil)
	u2 := NewBlock(g, 4, 12, nil)
	b := NewBlockWithUncles(a, 5, 20, nil, []*Block{u1, u2})

	uncles := b.Uncles()
	if got, want := len(uncles), 2; got != want {
		t.Fatalf("uncles len: got=%d want=%d", got, want)
	}
	uncles[0] = nil
	if b.Uncles()[0] == nil {
		t.Fatal("Uncles() should return a defensive copy")
	}

	ids := b.UncleIDs()
	if got, want := len(ids), 2; got != want {
		t.Fatalf("uncle ids len: got=%d want=%d", got, want)
	}
	if ids[0] != u1.ID() || ids[1] != u2.ID() {
		t.Fatalf("unexpected uncle ids: got=%v want=[%d %d]", ids, u1.ID(), u2.ID())
	}
}
