package tasks

import (
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

type stubEndpoint struct {
	id      int
	region  int
	compact bool

	received []MessageTask
	sentNext int
	flowed   int
}

func (s *stubEndpoint) ID() int                            { return s.id }
func (s *stubEndpoint) Region() int                        { return s.region }
func (s *stubEndpoint) SupportsCompactBlockRelay() bool    { return s.compact }
func (s *stubEndpoint) SendNextBlockMessage()              { s.sentNext++ }
func (s *stubEndpoint) ReceiveMessage(message MessageTask) { s.received = append(s.received, message) }
func (s *stubEndpoint) RecordFlowBlock(_ Endpoint, _ *core.Block, _ core.SimTime) {
	s.flowed++
}

func TestMessageTaskIntervalsAndDelivery(t *testing.T) {
	net := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)
	from := &stubEndpoint{id: 1, region: 0}
	to := &stubEndpoint{id: 2, region: 1}
	block := core.GenesisBlock(1, nil)

	inv := NewInvMessageTask(from, to, block, net)
	if got, want := inv.Interval(), core.SimTime(15); got != want {
		t.Fatalf("inv interval: got=%d want=%d", got, want)
	}
	inv.Run()
	if len(to.received) != 1 {
		t.Fatalf("expected 1 received message, got=%d", len(to.received))
	}

	delay := core.SimTime(20)
	bm := NewBlockMessageTask(from, to, block, delay, net)
	if got, want := bm.Interval(), core.SimTime(25); got != want {
		t.Fatalf("block interval: got=%d want=%d", got, want)
	}
	bm.Run()
	if from.sentNext != 1 {
		t.Fatalf("expected sendNext callback once, got=%d", from.sentNext)
	}
	if from.flowed != 1 {
		t.Fatalf("expected flow callback once, got=%d", from.flowed)
	}
}

func TestMiningTaskRunCallback(t *testing.T) {
	called := false
	task := NewMiningTask(12, func() { called = true })
	if got, want := task.Interval(), core.SimTime(12); got != want {
		t.Fatalf("interval: got=%d want=%d", got, want)
	}
	task.Run()
	if !called {
		t.Fatal("mining callback was not called")
	}
}
