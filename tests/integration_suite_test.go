package tests

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/teiyou416/simblock_go/core"
	"github.com/teiyou416/simblock_go/engine"
	"github.com/teiyou416/simblock_go/network"
	"github.com/teiyou416/simblock_go/node/consensus"
)

type suiteTask struct {
	interval core.SimTime
	run      func()
}

func (t suiteTask) Interval() core.SimTime { return t.interval }
func (t suiteTask) Run()                   { t.run() }

func testNetworkModel() *network.Model {
	return network.NewModel(
		[][]core.SimTime{
			{1, 5},
			{5, 1},
		},
		[]uint64{100_000, 100_000},
		[]uint64{100_000, 100_000},
	)
}

func TestIntegratedSuite(t *testing.T) {
	t.Run("timer_runs_all_scheduled_tasks", func(t *testing.T) {
		timer := engine.NewTimer()
		executed := make(map[string]int, 3)

		timer.PutTaskAt(suiteTask{run: func() { executed["a"]++ }}, 10)
		timer.PutTaskAt(suiteTask{run: func() { executed["b"]++ }}, 10)
		timer.PutTaskAt(suiteTask{run: func() { executed["c"]++ }}, 5)
		timer.RunUntilEmpty()

		if timer.Len() != 0 {
			t.Fatalf("timer queue should be empty, got=%d", timer.Len())
		}
		if timer.CurrentTime() != 10 {
			t.Fatalf("unexpected current time: got=%d want=10", timer.CurrentTime())
		}
		if executed["a"] != 1 || executed["b"] != 1 || executed["c"] != 1 {
			t.Fatalf("unexpected executed map: %+v", executed)
		}
	})

	t.Run("pow_minting_interval_non_negative", func(t *testing.T) {
		pow := consensus.NewPoW(consensus.PoWConfig{
			InitialDifficulty: 100,
			TargetInterval:    10,
			AdjustmentWindow:  3,
			DifficultyMode:    consensus.DifficultyStatic,
		}, rand.New(rand.NewSource(7)))

		task := pow.Minting(pow.GenesisBlock(1), 1, 1000)
		if task == nil {
			t.Fatal("Minting returned nil task")
		}
		if task.Interval() < 0 {
			t.Fatalf("minting interval should be non-negative, got=%d", task.Interval())
		}
	})

	t.Run("simulator_smoke_outputs_expected_files", func(t *testing.T) {
		timer := engine.NewTimer()
		outDir := t.TempDir()

		sim := engine.NewSimulator(engine.SimulatorConfig{
			NumNodes:           16,
			TargetInterval:     50,
			EndTime:            8_000,
			BlockSize:          535000,
			OutputDir:          outDir,
			RandomSeed:         10,
			ConnectionsPerNode: 2,
		}, timer, testNetworkModel())

		stats, err := sim.Run()
		if err != nil {
			t.Fatalf("simulator run failed: %v", err)
		}
		if stats.AcceptedBlocks == 0 {
			t.Fatal("expected accepted blocks > 0")
		}

		for _, name := range []string{"output.json", "static.json", "metrics.json"} {
			if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
				t.Fatalf("missing output file %s: %v", name, err)
			}
		}
	})
}
