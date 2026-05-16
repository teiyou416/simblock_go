package engine

import (
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/node"
	"github.com/teiyou416/simblock_go/node/consensus"
)

func TestCollectStatsCountsObservedBlocksOffCanonicalChainAsOrphans(t *testing.T) {
	n := node.New(1, 0)
	genesis := core.GenesisBlock(1, consensus.PoWData{
		Difficulty:      0,
		TotalDifficulty: 0,
		NextDifficulty:  1,
	})
	main := core.NewBlock(genesis, 1, 10, consensus.PoWData{
		Difficulty:      1,
		TotalDifficulty: 1,
		NextDifficulty:  1,
	})
	stale := core.NewBlock(genesis, 2, 11, consensus.PoWData{
		Difficulty:      1,
		TotalDifficulty: 1,
		NextDifficulty:  1,
	})

	if !n.ReceiveBlock(genesis) {
		t.Fatal("expected genesis accepted")
	}
	if !n.ReceiveBlock(main) {
		t.Fatal("expected main block accepted")
	}

	sim := &Simulator{
		nodes: []*node.Node{n},
		seenBlocks: map[uint64]*core.Block{
			genesis.ID(): genesis,
			main.ID():    main,
			stale.ID():   stale,
		},
	}

	stats := sim.collectStats()
	if got, want := stats.OrphanBlocks, 1; got != want {
		t.Fatalf("OrphanBlocks: got=%d want=%d", got, want)
	}
	if got, want := stats.OrphanRate, 1.0/3.0; got != want {
		t.Fatalf("OrphanRate: got=%f want=%f", got, want)
	}
}
