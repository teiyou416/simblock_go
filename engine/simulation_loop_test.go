package engine_test

import (
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node"
)

func TestMinimalEngineDomainLoop(t *testing.T) {
	timer := engine.NewTimer()
	net := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)

	miner := node.NewWithHashPower(1, 0, 2000)
	receiver := node.NewWithHashPower(2, 1, 1000)
	miner.BindEnvironment(timer, net)
	receiver.BindEnvironment(timer, net)
	miner.AddNeighbor(receiver)
	receiver.AddNeighbor(miner)

	genesis := core.GenesisBlock(miner.ID(), nil)
	miner.ReceiveBlock(genesis)
	receiver.ReceiveBlock(genesis)

	mined := core.NewBlock(miner.Tip(), miner.ID(), timer.CurrentTime(), nil)
	miner.ReceiveBlock(mined)

	timer.RunUntilEmpty()

	if got := miner.Tip().Height(); got == 0 {
		t.Fatalf("miner tip height: got=%d want>0", got)
	}
	if got := receiver.Tip().Height(); got == 0 {
		t.Fatalf("receiver tip height: got=%d want>0", got)
	}
	if miner.Tip().ID() != receiver.Tip().ID() {
		t.Fatalf("tips diverged: miner=%d receiver=%d", miner.Tip().ID(), receiver.Tip().ID())
	}
}
