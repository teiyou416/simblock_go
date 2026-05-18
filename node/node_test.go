package node_test

import (
	"math/rand"
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node"
	"github.com/teiyou416/simblock_go/node/consensus"
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

func TestSetBlockSize(t *testing.T) {
	n := node.NewWithHashPower(1, 0, 10)
	defaultSize := n.BlockSize()
	if defaultSize == 0 {
		t.Fatal("default block size should be > 0")
	}

	n.SetBlockSize(1234)
	if got, want := n.BlockSize(), uint64(1234); got != want {
		t.Fatalf("BlockSize() after SetBlockSize: got=%d want=%d", got, want)
	}

	n.SetBlockSize(0)
	if got, want := n.BlockSize(), uint64(1234); got != want {
		t.Fatalf("BlockSize() should not change on zero input: got=%d want=%d", got, want)
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

func TestForkChoiceHeaviestVsGHOST(t *testing.T) {
	buildChild := func(parent *core.Block, minter int, ts core.SimTime) *core.Block {
		parentData, ok := consensus.PoWDataFromBlock(parent)
		if !ok {
			t.Fatal("missing parent PoW data")
		}
		return core.NewBlock(parent, minter, ts, consensus.PoWData{
			Difficulty:      1,
			TotalDifficulty: parentData.TotalDifficulty + 1,
			NextDifficulty:  1,
		})
	}
	newNode := func(forkChoice string) *node.Node {
		n := node.NewWithHashPower(1, 0, 1000)
		n.SetForkChoice(forkChoice)
		n.SetConsensus(consensus.NewPoW(consensus.PoWConfig{
			InitialDifficulty: 1,
			TargetInterval:    10,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, rand.New(rand.NewSource(7))))
		return n
	}

	genesis := core.GenesisBlock(1, consensus.PoWData{
		Difficulty:      0,
		TotalDifficulty: 0,
		NextDifficulty:  1,
	})
	a1 := buildChild(genesis, 2, 10)
	a2 := buildChild(a1, 3, 20)
	a3 := buildChild(a2, 4, 30)
	b1 := buildChild(genesis, 5, 11)
	b2 := buildChild(b1, 6, 21)
	b3 := buildChild(b1, 7, 22)
	b4 := buildChild(b1, 8, 23)
	b5 := buildChild(b1, 9, 24)
	blocks := []*core.Block{genesis, a1, a2, a3, b1, b2, b3, b4, b5}

	nHeaviest := newNode(string(node.ForkChoiceHeaviest))
	for _, b := range blocks {
		nHeaviest.ReceiveBlock(b)
	}
	if got, want := nHeaviest.Tip().ID(), a3.ID(); got != want {
		t.Fatalf("heaviest tip mismatch: got=%d want=%d", got, want)
	}

	nGhost := newNode(string(node.ForkChoiceGHOST))
	for _, b := range blocks {
		nGhost.ReceiveBlock(b)
	}
	if got, want := nGhost.Tip().ID(), b2.ID(); got != want {
		t.Fatalf("ghost tip mismatch: got=%d want=%d", got, want)
	}
}

func TestGHOSTRejectsBlockWithTooManyUncles(t *testing.T) {
	newPoWNode := func() *node.Node {
		n := node.NewWithHashPower(1, 0, 1000)
		n.SetForkChoice(string(node.ForkChoiceGHOST))
		n.SetConsensus(consensus.NewPoW(consensus.PoWConfig{
			InitialDifficulty: 1,
			TargetInterval:    10,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, rand.New(rand.NewSource(11))))
		return n
	}
	buildChild := func(parent *core.Block, minter int, ts core.SimTime, uncles []*core.Block) *core.Block {
		parentData, ok := consensus.PoWDataFromBlock(parent)
		if !ok {
			t.Fatal("missing parent PoW data")
		}
		return core.NewBlockWithUncles(parent, minter, ts, consensus.PoWData{
			Difficulty:      1,
			TotalDifficulty: parentData.TotalDifficulty + 1,
			NextDifficulty:  1,
		}, uncles)
	}

	g := core.GenesisBlock(1, consensus.PoWData{Difficulty: 0, TotalDifficulty: 0, NextDifficulty: 1})
	a1 := buildChild(g, 2, 10, nil)
	u1 := buildChild(g, 3, 11, nil)
	u2 := buildChild(g, 4, 12, nil)
	u3 := buildChild(g, 5, 13, nil)
	tooMany := buildChild(a1, 6, 20, []*core.Block{u1, u2, u3})

	n := newPoWNode()
	if !n.ReceiveBlock(g) {
		t.Fatal("expected genesis accepted")
	}
	n.ReceiveBlock(a1)
	n.ReceiveBlock(u1)
	n.ReceiveBlock(u2)
	n.ReceiveBlock(u3)
	if n.Tip() == nil {
		t.Fatal("tip should not be nil after seeding known blocks")
	}
	if n.ReceiveBlock(tooMany) {
		t.Fatal("expected block with >2 uncles to be rejected")
	}
}

func TestGHOSTRejectsReusedUncle(t *testing.T) {
	newPoWNode := func() *node.Node {
		n := node.NewWithHashPower(1, 0, 1000)
		n.SetForkChoice(string(node.ForkChoiceGHOST))
		n.SetConsensus(consensus.NewPoW(consensus.PoWConfig{
			InitialDifficulty: 1,
			TargetInterval:    10,
			AdjustmentWindow:  1,
			DifficultyMode:    consensus.DifficultyStatic,
		}, rand.New(rand.NewSource(13))))
		return n
	}
	buildChild := func(parent *core.Block, minter int, ts core.SimTime, uncles []*core.Block) *core.Block {
		parentData, ok := consensus.PoWDataFromBlock(parent)
		if !ok {
			t.Fatal("missing parent PoW data")
		}
		return core.NewBlockWithUncles(parent, minter, ts, consensus.PoWData{
			Difficulty:      1,
			TotalDifficulty: parentData.TotalDifficulty + 1,
			NextDifficulty:  1,
		}, uncles)
	}

	g := core.GenesisBlock(1, consensus.PoWData{Difficulty: 0, TotalDifficulty: 0, NextDifficulty: 1})
	a1 := buildChild(g, 2, 10, nil)
	u1 := buildChild(g, 3, 11, nil)
	a2 := buildChild(a1, 4, 20, []*core.Block{u1})
	a3Reuse := buildChild(a2, 5, 30, []*core.Block{u1})

	n := newPoWNode()
	if !n.ReceiveBlock(g) {
		t.Fatal("expected genesis accepted")
	}
	n.ReceiveBlock(a1)
	n.ReceiveBlock(u1)
	if !n.ReceiveBlock(a2) {
		t.Fatal("expected first block including uncle to be accepted")
	}
	if n.ReceiveBlock(a3Reuse) {
		t.Fatal("expected reused uncle to be rejected")
	}
}
