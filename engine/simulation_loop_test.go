package engine_test

import (
	"testing"

	"github.com/teiyou416/simblock_go/block"
	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node"
)

type mineAndBroadcastTask struct {
	interval core.SimTime
	timer    *engine.Timer
	net      *network.Model
	miner    *node.Node
	peer     *node.Node
}

func (t *mineAndBroadcastTask) Interval() core.SimTime { return t.interval }

func (t *mineAndBroadcastTask) Run() {
	mined := t.miner.MintBlock(t.timer.CurrentTime())
	delay := t.net.Latency(t.miner.Region(), t.peer.Region())
	t.timer.PutTaskAt(&deliverBlockTask{
		block: mined,
		to:    t.peer,
	}, t.timer.CurrentTime()+delay)
}

type deliverBlockTask struct {
	block *block.Block
	to    *node.Node
}

func (t *deliverBlockTask) Interval() core.SimTime { return 0 }

func (t *deliverBlockTask) Run() {
	t.to.ReceiveBlock(t.block)
}

func TestMinimalEngineDomainLoop(t *testing.T) {
	timer := engine.NewTimer()
	net := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100, 100},
		[]uint64{100, 100},
	)

	miner := node.New(1, 0)
	receiver := node.New(2, 1)
	genesis := block.Genesis(miner.ID())
	miner.ReceiveBlock(genesis)
	receiver.ReceiveBlock(genesis)

	timer.PutTask(&mineAndBroadcastTask{
		interval: 10,
		timer:    timer,
		net:      net,
		miner:    miner,
		peer:     receiver,
	})

	timer.RunUntilEmpty()

	if got := miner.Tip().Height(); got != 1 {
		t.Fatalf("miner tip height: got=%d want=1", got)
	}
	if got := receiver.Tip().Height(); got != 1 {
		t.Fatalf("receiver tip height: got=%d want=1", got)
	}
	if miner.Tip().ID() != receiver.Tip().ID() {
		t.Fatalf("tips diverged: miner=%d receiver=%d", miner.Tip().ID(), receiver.Tip().ID())
	}
	if got, want := timer.CurrentTime(), core.SimTime(15); got != want {
		t.Fatalf("current time: got=%d want=%d", got, want)
	}
}
