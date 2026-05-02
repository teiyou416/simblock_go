package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
)

func TestSimulatorRunAndOutputFiles(t *testing.T) {
	timer := engine.NewTimer()
	netModel := network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)

	outDir := t.TempDir()
	sim := engine.NewSimulator(engine.SimulatorConfig{
		NumNodes:           8,
		TargetInterval:     50,
		EndTime:            10_000,
		BlockSize:          535000,
		OutputDir:          outDir,
		RandomSeed:         10,
		ConnectionsPerNode: 2,
	}, timer, netModel)

	stats, err := sim.Run()
	if err != nil {
		t.Fatalf("Run() err=%v", err)
	}

	if stats.AcceptedBlocks == 0 {
		t.Fatal("expected accepted blocks > 0")
	}

	for _, name := range []string{"output.json", "static.json", "metrics.json"} {
		p := filepath.Join(outDir, name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected output file %s: %v", p, err)
		}
	}
}
