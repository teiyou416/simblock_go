package engine

import (
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/network"
)

func TestSetupAppliesConfiguredBlockSizeToNodes(t *testing.T) {
	netModel := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)

	sim := NewSimulator(SimulatorConfig{
		NumNodes:           8,
		TargetInterval:     50,
		EndTime:            1_000,
		BlockSize:          1234,
		RandomSeed:         10,
		ConnectionsPerNode: 2,
	}, NewTimer(), netModel)

	if err := sim.Setup(); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	for _, n := range sim.Nodes() {
		if got, want := n.BlockSize(), uint64(1234); got != want {
			t.Fatalf("node %d BlockSize(): got=%d want=%d", n.ID(), got, want)
		}
	}
}

func TestSetupAppliesConfiguredForkChoiceToNodes(t *testing.T) {
	netModel := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)

	sim := NewSimulator(SimulatorConfig{
		NumNodes:           8,
		TargetInterval:     50,
		EndTime:            1_000,
		BlockSize:          1234,
		ForkChoice:         "ghost",
		RandomSeed:         10,
		ConnectionsPerNode: 2,
	}, NewTimer(), netModel)

	if err := sim.Setup(); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	for _, n := range sim.Nodes() {
		if got, want := n.ForkChoice(), "ghost"; got != want {
			t.Fatalf("node %d ForkChoice(): got=%s want=%s", n.ID(), got, want)
		}
	}
}
